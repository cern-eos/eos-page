package store

import (
	"embed"
	"encoding/json"
	"strings"
)

//go:embed seeddata
var seedFS embed.FS

func seedBytes() []byte {
	b, err := seedFS.ReadFile("seeddata/index.json")
	if err != nil {
		panic(err)
	}
	return b
}

type seedFile struct {
	Workshops []seedWorkshop `json:"workshops"`
	Talks     []Talk         `json:"talks"`
	Docs      []Doc          `json:"docs"`
}

type seedWorkshop struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Year        int    `json:"year"`
	Edition     int    `json:"edition"`
	Start       string `json:"start"`
	End         string `json:"end"`
	Location    string `json:"location"`
	Room        string `json:"room"`
	Kind        string `json:"kind"`
	URL         string `json:"url"`
	Description string `json:"description"`
}

func (s *Store) seedIfEmpty() error {
	var n int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM pages`).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return nil
	}

	settings := map[string]string{
		"hero_video":         "ttSjYYBOlsM",
		"hero_kicker":        "CERN storage technology used at the Large Hadron Collider (LHC)",
		"hero_title":         "EOS Open Storage",
		"hero_lede":          "Open-source disk storage for interactive and batch analysis - built at CERN, used worldwide.",
		"latest_version":     "5.5.1",
		"latest_version_url": "https://eos-docs.web.cern.ch/diopside/releases/diopside-release.html#v5-5-1-diopside",
		"workshop_label":     "Workshop '26",
		"workshop_url":       "https://indico.cern.ch/event/1622471/",
		"install_url":        "https://eos-docs.web.cern.ch/diopside/manual/getting-started.html",
		"stat_volume":        "1.1 EB",
		"stat_volume_label":  "Storage volume at CERN",
		"stat_io":            "1–2 TB/s",
		"stat_io_label":      "IO",
		"stat_disks":         "100k",
		"stat_disks_label":   "Hard disks",
		"stat_files":         "8 000 M",
		"stat_files_label":   "Files",
		"stat_clients":       "30k",
		"stat_clients_label": "Clients",
		"contact_email":      "eos-support@cern.ch",
		"address":            "CERN Storage & Data Management Group, Esplanade des Particules 1, 1211 Geneva, Switzerland",
		"github":             "https://github.com/cern-eos/eos",
		"gitlab":             "https://gitlab.cern.ch/dss/eos",
		"community":          "https://eos-community.web.cern.ch/",
		"jira":               "https://its.cern.ch/jira/secure/RapidBoard.jspa?rapidView=4511&projectKey=EOS",
		"ci":                 "https://gitlab.cern.ch/dss/eos/pipelines",
		"docs_url":           "https://eos-docs.web.cern.ch/diopside/",
		"release_notes_url":  "https://eos-docs.web.cern.ch/diopside/releases/diopside-release.html",
		"rpm_url":            "https://storage-ci.web.cern.ch/storage-ci/eos/diopside/tag/el-9/x86_64/",
		"status_board":       "https://cern.service-now.com/service-portal?id=service_status_board&area=IT",
		"control_tower":      "https://monit-grafana.cern.ch/d/baff3c33-decb-4b91-a6bf-c0ba84bdcbe4/eos-user-monitoring?orgId=22&from=now-24h&to=now&timezone=browser&var-cluster=$__all&var-HTTP=$__all&var-GRIDFPT=$__all&var-XROOTD=$__all&var-FUSE=$__all",
		"presentations_url":  "https://cernbox.cern.ch/index.php/s/Kl0hxpeA5bFQ4Ho?path=%2Fpresentations",
		"publications_url":   "https://cernbox.cern.ch/index.php/s/Kl0hxpeA5bFQ4Ho?path=%2Fpublications",
		"indico_event_ids":   "1622471,1483930,1353101,1227241,1103358,985953,862873,775181,656157",
		"docs_base":          "https://eos-docs.web.cern.ch/diopside/",
	}
	for k, v := range settings {
		if err := s.SetSetting(k, v); err != nil {
			return err
		}
	}

	pages := []Page{
		{ID: "about", Title: "About EOS", Body: "EOS provides a service for storing large amounts of physics data and user files, with a focus on interactive and batch analysis. It started at CERN in 2010 as disk-only storage for LHC analysis and is now the foundation of CERNBox, CTA tape workflows, and dozens of WLCG deployments."},
		{ID: "tech", Title: "Design & architecture", Body: techPageBody},
		{ID: "support", Title: "Support", Body: "Operational questions for the CERN service go to eos-support@cern.ch. Software issues belong in the EOS JIRA tracker. Community discussion and site reports happen at the yearly EOS workshop and on the community pages."},
		{ID: "roadmap", Title: "Roadmap", Body: roadmapPageBody},
	}
	for _, p := range pages {
		if err := s.UpsertPage(p); err != nil {
			return err
		}
	}

	features := []Card{
		{ID: "f-flex", Kind: "feature", Sort: 1, Visible: true, Title: "Flexible", Body: "A storage system for central data recording, analysis, and processing - from a single node to an exabyte service."},
		{ID: "f-scale", Kind: "feature", Sort: 2, Visible: true, Title: "Adaptable and scalable", Body: "Thousands of clients with random remote I/O. Native protocols: XRootD, HTTP/WebDAV, gRPC, FUSE, and CIFS."},
		{ID: "f-cern", Kind: "feature", Sort: 3, Visible: true, Title: "Built for CERN scale", Body: "Designed for high capacity and low latency. The CERN deployment is in the exabyte class for disk plus tape."},
		{ID: "f-sec", Kind: "feature", Sort: 4, Visible: true, Title: "Security", Body: "KRB5, X.509, OIDC, shared secret, JWT, and EOS tokens. Virtual identities map every client onto a uid/gid plus roles."},
		{ID: "f-sync", Kind: "feature", Sort: 5, Visible: true, Title: "Sync & share", Body: "EOS is the storage backend for CERNBox: web portal, desktop sync, and at least a terabyte of personal space."},
		{ID: "f-tape", Kind: "feature", Sort: 6, Visible: true, Title: "Tape storage", Body: "CTA (CERN Tape Archive) uses EOS as the user-facing disk cache in front of the tape infrastructure."},
	}
	components := techComponentCards()
	resources := []Card{
		{ID: "r-docs", Kind: "resource", Sort: 1, Visible: true, Title: "Documentation", Href: "https://eos-docs.web.cern.ch/diopside/",
			Body: "Install, configure, and operate EOS 5 Diopside - architecture, manual, FAQ, and release notes."},
		{ID: "r-search", Kind: "resource", Sort: 2, Visible: true, Title: "Workshop talks", Href: "/search",
			Body: "Search every EOS workshop contribution since 2018: slides, recordings, and abstracts."},
		{ID: "r-pres", Kind: "resource", Sort: 3, Visible: true, Title: "Presentations", Href: "https://cernbox.cern.ch/index.php/s/Kl0hxpeA5bFQ4Ho?path=%2Fpresentations",
			Body: "A curated collection of conference talks about EOS."},
		{ID: "r-pubs", Kind: "resource", Sort: 4, Visible: true, Title: "Publications", Href: "/resources#publications",
			Body: "Papers on EOS design, operations, LHC data handling, and XRootD."},
		{ID: "r-rel", Kind: "resource", Sort: 5, Visible: true, Title: "Release notes", Href: "https://eos-docs.web.cern.ch/diopside/releases/diopside-release.html",
			Body: "Diopside 5.x notes. Current stable line: 5.5.1 (August 2026)."},
		searchCommitsCard(),
	}
	services := serviceCards()
	collabs := []Card{
		{ID: "col-jrc", Kind: "collab", Sort: 1, Visible: true, Title: "Joint Research Centre", Href: "https://joint-research-centre.ec.europa.eu/", Body: "European Commission big-data platform"},
		{ID: "col-xrd", Kind: "collab", Sort: 2, Visible: true, Title: "XRootD", Href: "https://xrootd.slac.stanford.edu/", Body: "Native protocol and server framework"},
	}
	links := []Card{
		{ID: "l-gh", Kind: "link", Sort: 1, Visible: true, Title: "GitHub", Href: "https://github.com/cern-eos/eos", Body: "Public source and issues mirror"},
		{ID: "l-gl", Kind: "link", Sort: 2, Visible: true, Title: "GitLab", Href: "https://gitlab.cern.ch/dss/eos", Body: "Canonical CERN repository"},
		{ID: "l-com", Kind: "link", Sort: 3, Visible: true, Title: "Community", Href: "https://eos-community.web.cern.ch/", Body: "Sites and operators"},
		{ID: "l-jira", Kind: "link", Sort: 4, Visible: true, Title: "JIRA", Href: "https://its.cern.ch/jira/secure/RapidBoard.jspa?rapidView=4511&projectKey=EOS", Body: "Bug and feature tracker"},
		{ID: "l-ci", Kind: "link", Sort: 5, Visible: true, Title: "CI", Href: "https://gitlab.cern.ch/dss/eos/pipelines", Body: "GitLab pipelines"},
		{ID: "l-rpm", Kind: "link", Sort: 6, Visible: true, Title: "RPMs", Href: "https://storage-ci.web.cern.ch/storage-ci/eos/diopside/tag/el-9/x86_64/", Body: "Diopside packages"},
	}
	for _, group := range [][]Card{features, components, resources, services, collabs, links, helpCards()} {
		for _, c := range group {
			if err := s.UpsertCard(c); err != nil {
				return err
			}
		}
	}

	news := []News{
		{ID: "n-ws27", Sort: 0, Visible: true, Title: "Stay tuned for the next EOS workshop in spring 2027", DateLabel: "22 September 2026",
			Href: "/workshops",
			Body: "The tenth workshop is behind us. The next EOS workshop is planned for spring 2027 - dates and the Indico page will be announced here."},
		{ID: "n-ws26", Sort: 1, Visible: true, Title: "10th EOS Workshop at CERN", DateLabel: "9–11 March 2026",
			Href: "https://indico.cern.ch/event/1622471/",
			Body: "Developers, sites, and users met at CERN for the tenth workshop: roadmap toward EOS 6, I/O shaping, S3 and NFS research, operations, and worldwide deployments. Slides and recordings are in Indico."},
		{ID: "n-551", Sort: 2, Visible: true, Title: "EOS 5.5.1 Diopside", DateLabel: "26 August 2026",
			Href: "https://eos-docs.web.cern.ch/diopside/releases/diopside-release.html#v5-5-1-diopside",
			Body: "Latest stable release. Built on XRootD 5/6, with HTTP-TPC integrity, digest headers, and a long list of MGM/FST/FUSE fixes."},
		{ID: "n-exa", Sort: 3, Visible: true, Title: "CERN passed an exabyte of disk storage", DateLabel: "29 September 2023",
			Href: "https://home.cern/news/news/computing/exabyte-disk-storage-cern",
			Body: "Most of that capacity is managed by EOS services."},
	}
	for _, n := range news {
		if err := s.UpsertNews(n); err != nil {
			return err
		}
	}

	people := communityPeople()
	for _, p := range people {
		if err := s.UpsertPerson(p); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) ensureIndex() error {
	var n int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM talks`).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	return s.LoadSeedIndex()
}

func (s *Store) LoadSeedIndex() error {
	var payload seedFile
	if err := json.Unmarshal(seedBytes(), &payload); err != nil {
		return err
	}
	for _, w := range payload.Workshops {
		if err := s.UpsertWorkshop(Workshop{
			ID: w.ID, Title: w.Title, Year: w.Year, Edition: w.Edition,
			Start: w.Start, End: w.End, Location: w.Location, Room: w.Room,
			Kind: w.Kind, URL: w.URL, Description: w.Description,
			Sort: 3000 - w.Year, Visible: true,
		}); err != nil {
			return err
		}
	}
	if err := s.ReplaceTalks(payload.Talks); err != nil {
		return err
	}
	for i := range payload.Docs {
		payload.Docs[i].Body = strings.TrimSpace(payload.Docs[i].Body)
		payload.Docs[i].Sort = i
	}
	return s.ReplaceDocs(payload.Docs)
}

const roadmapPageBody = "The EOS development programme covers production improvements, architectural evolution and selected R&D. The aim is a policy-driven platform that caches, places and archives across disk, erasure coding and tape - and that can use data locality and energy-aware placement on future CERN computing infrastructures."

const techPageBody = "EOS was designed as a file storage system with low-latency access for physics analysis. The production release is Diopside - EOS 5 - with XRootD as the native protocol.\n\nXRootD adds what FTP, NFS and plain HTTP do not: strong authentication, a redirection protocol that separates metadata from data (and supports federations and error recovery), vector reads and writes for WAN latency, third-party copy with checksums at both ends, and wait / waitresp for asynchronous callbacks.\n\nA cluster is four services: the MGM (hierarchical namespace), MQ (asynchronous MGM–FST messaging), FSTs (file storage), and QuarkDB (highly available key-value persistency for metadata). HTTP(S) is a native XrdHttp plugin. S3, CIFS and SFTP arrive through gateways; eosxd mounts the same feature set as a filesystem."

func techComponentCards() []Card {
	return []Card{
		{ID: "c-mgm", Kind: "component", Sort: 1, Visible: true, Title: "MGM · namespace", Meta: "Management server",
			Body: "Hierarchical namespace and metadata access. The EOS 5 namespace runs on a single active node - it is not sharded - with standby hosts that take over if the master is lost. Metadata is an LRU in-memory cache with a write-back queue into QuarkDB."},
		{ID: "c-mq", Kind: "component", Sort: 2, Visible: true, Title: "MQ · messaging", Meta: "Message queue",
			Body: "Asynchronous messaging between MGM and FST services. QDB and MQ have no startup dependencies; the MGM needs both before it can run, and each FST needs the MGM, QDB and MQ."},
		{ID: "c-qdb", Kind: "component", Sort: 3, Visible: true, Title: "QuarkDB · persistency", Meta: "RAFT KV store", Href: "https://github.com/gbitzes/QuarkDB",
			Body: "High-available transactional key-value store using the RAFT consensus algorithm and a Redis subset. A typical cluster is three nodes electing a leader. Each node stores hashes, sets, strings, leases and pub/sub in RocksDB; replication uses RAFT journals."},
		{ID: "c-fst", Kind: "component", Sort: 4, Visible: true, Title: "FST · storage server", Meta: "File storage",
			Body: "Holds file data on local disks with replica and erasure-coded layouts - for example raid6 on 12 HDDs. Filesystems are identified by path, host, an internal ID (1–65534) and a UUID stored on the disk. Checksums, drain and consistency repair live here."},
		{ID: "c-cli", Kind: "component", Sort: 5, Visible: true, Title: "Clients & protocols", Meta: "eos · eosxd · XrdCl",
			Body: "The eos shell for users and operators. Native XRootD clients: xrdcp, xrd, and the XrdCl C++ library. eosxd FUSE mounts EOS with the same authentication and tokens. HTTP(S)/WebDAV is an XrdHttp plugin; S3 (MinIO), CIFS (Samba) and SFTP (sshfs) go through gateways."},
	}
}

func (s *Store) ensureRoadmapContent() error {
	return s.UpsertPage(Page{ID: "roadmap", Title: "Roadmap", Body: roadmapPageBody})
}

func communityPeople() []Person {
	core := []Person{
		{ID: "p-ajp", Sort: 1, Visible: true, Name: "Andreas-Joachim Peters", Role: "Project Lead & Core Developer", Email: "andreas.joachim.peters@cern.ch"},
		{ID: "p-david", Sort: 2, Visible: true, Name: "David Smith", Role: "Core developer · CERN staff", Email: "david.smith@cern.ch"},
		{ID: "p-elvin", Sort: 3, Visible: true, Name: "Elvin Alin Sindrilaru", Role: "Core developer · Operations · CERN staff", Email: "elvin.alin.sindrilaru@cern.ch"},
		{ID: "p-cedric", Sort: 4, Visible: true, Name: "Cedric Caffy", Role: "Core developer · Operations · CERN staff", Email: "cedric.caffy@cern.ch"},
		{ID: "p-luis", Sort: 5, Visible: true, Name: "Luis Obis", Role: "Core developer · CERN staff", Email: "luis.obis@cern.ch"},
		{ID: "p-diogo", Sort: 6, Visible: true, Name: "Diogo Mattiolo", Role: "Core developer · Operations · CERN staff", Email: "diogo.mattiolo@cern.ch"},
		{ID: "p-gianmaria", Sort: 7, Visible: true, Name: "Gianmaria Del Monte", Role: "Core developer · CERN staff", Email: "gianmaria.del.monte@cern.ch"},
		{ID: "p-amadio", Sort: 8, Visible: true, Name: "Guilherme Amadio", Role: "Core developer · Operations · CERN staff", Email: "guilherme.amadio@cern.ch"},
		{ID: "p-octavian", Sort: 9, Visible: true, Name: "Octavian-Mihai Matei", Role: "Core Developer - Operations", Email: "octavian-mihai.matei@cern.ch"},
	}
	ops := []Person{
		{ID: "p-luca", Sort: 20, Visible: true, Name: "Luca Mascetti", Role: "Physics & Data Services Section Lead & openlab CTO", Email: "luca.mascetti@cern.ch"},
		{ID: "p-ruhi", Sort: 21, Visible: true, Name: "Ruhi Choudhury", Role: "Operations Lead & openlab", Email: "ruhi.choudhury@cern.ch"},
	}
	return append(core, ops...)
}

func (s *Store) ensureCommunityRoster() error {
	for _, id := range []string{"col-aarnet", "col-comtrade"} {
		if err := s.DeleteCard(id); err != nil && err != ErrNotFound {
			return err
		}
	}
	if err := s.DeletePerson("p-abhi"); err != nil && err != ErrNotFound {
		return err
	}
	for _, p := range communityPeople() {
		if err := s.UpsertPerson(p); err != nil {
			return err
		}
	}
	return nil
}

func serviceCards() []Card {
	return []Card{
		{ID: "s-tower", Kind: "service", Sort: 1, Visible: true, Title: "Control Tower", Href: "https://monit-grafana.cern.ch/d/baff3c33-decb-4b91-a6bf-c0ba84bdcbe4/eos-user-monitoring?orgId=22&from=now-24h&to=now&timezone=browser&var-cluster=$__all&var-HTTP=$__all&var-GRIDFPT=$__all&var-XROOTD=$__all&var-FUSE=$__all",
			Meta: "/static/media/control-tower.jpg",
			Body: "Live Grafana view of CERN EOS instances."},
		{ID: "s-orbit", Kind: "service", Sort: 2, Visible: true, Title: "EOS Orbit", Href: "https://eos-orbit.cern.ch",
			Meta: "/static/media/eos-orbit.jpg",
			Body: "Live map of EOS instances - nodes, traffic and placement."},
		{ID: "s-box", Kind: "service", Sort: 3, Visible: true, Title: "CERNBox", Href: "https://cernbox.web.cern.ch/",
			Meta: "/static/media/cernbox-web.jpg",
			Body: "Sync and share on top of EOS."},
		{ID: "s-swan", Kind: "service", Sort: 4, Visible: true, Title: "SWAN", Href: "https://swan.web.cern.ch/",
			Meta: "/static/media/swan-web.jpg",
			Body: "Notebooks that read and write EOS home and project spaces."},
		{ID: "s-cta", Kind: "service", Sort: 5, Visible: true, Title: "CTA", Href: "https://cta.web.cern.ch/",
			Meta: "/static/media/cta-web.jpg",
			Body: "CERN Tape Archive - disk cache plus tape."},
		itStatusCard(),
	}
}

func searchCommitsCard() Card {
	return Card{
		ID: "r-commits", Kind: "resource", Sort: 6, Visible: true, Title: "Search commits",
		Href: "/commits",
		Body: "Live git log of the EOS master branch: headings, messages, and authors.",
	}
}

func itStatusCard() Card {
	return Card{
		ID: "s-status", Kind: "service", Sort: 6, Visible: true, Title: "CERN IT status",
		Href: "https://cern.service-now.com/service-portal?id=service_status_board&area=IT",
		Body: "Planned interventions and incidents for CERN-hosted EOS services.",
	}
}

func (s *Store) ensureOrbitService() error {
	for _, c := range serviceCards() {
		if err := s.UpsertCard(c); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) ensureSearchCommits() error {
	var title, href string
	err := s.db.QueryRow(`SELECT title, href FROM cards WHERE id='r-status'`).Scan(&title, &href)
	if err == nil {
		low := strings.ToLower(title)
		if strings.Contains(low, "it status") || strings.Contains(href, "service-now") {
			if err := s.DeleteCard("r-status"); err != nil && err != ErrNotFound {
				return err
			}
		}
	}
	if err := s.UpsertCard(searchCommitsCard()); err != nil {
		return err
	}
	return s.UpsertCard(itStatusCard())
}

func (s *Store) ensurePublicationsCard() error {
	return s.UpsertCard(Card{
		ID: "r-pubs", Kind: "resource", Sort: 4, Visible: true, Title: "Publications",
		Href: "/resources#publications",
		Body: "Papers on EOS design, operations, LHC data handling, and XRootD.",
	})
}

func (s *Store) ensureTechContent() error {
	if err := s.UpsertPage(Page{ID: "tech", Title: "Design & architecture", Body: techPageBody}); err != nil {
		return err
	}
	for _, c := range techComponentCards() {
		if err := s.UpsertCard(c); err != nil {
			return err
		}
	}
	return s.UpsertCard(Card{
		ID: "h-arch", Kind: "help", Sort: 2, Visible: true, Title: "Architecture", Meta: "Design",
		Href: "https://eos-docs.web.cern.ch/diopside/architecture/index.html",
		Body: "Four services: MGM namespace, MQ messaging, FST data servers, QuarkDB persistency. Native protocol is XRootD.",
	})
}

func helpCards() []Card {
	return []Card{
		{ID: "h-intro", Kind: "help", Sort: 1, Visible: true, Title: "What is EOS?", Meta: "Introduction",
			Href: "https://eos-docs.web.cern.ch/diopside/introduction/index.html",
			Body: "Disk storage started at CERN in 2010 for LHC analysis. Today it underpins CERNBox, CTA, and WLCG sites."},
		{ID: "h-arch", Kind: "help", Sort: 2, Visible: true, Title: "Architecture", Meta: "Design",
			Href: "https://eos-docs.web.cern.ch/diopside/architecture/index.html",
			Body: "Four services: MGM namespace, MQ messaging, FST data servers, QuarkDB persistency. Native protocol is XRootD."},
		{ID: "h-start", Kind: "help", Sort: 3, Visible: true, Title: "Install in minutes", Meta: "Getting started",
			Href: "https://eos-docs.web.cern.ch/diopside/manual/getting-started.html",
			Body: "Alma/RHEL RPMs with eos daemon run, or a Helm chart that brings up MGM + QDB + FSTs."},
		{ID: "h-cli", Kind: "help", Sort: 4, Visible: true, Title: "Using the CLI", Meta: "Manual",
			Href: "https://eos-docs.web.cern.ch/diopside/manual/using.html",
			Body: "eos ls, find, cp, quota, recycle, and the rest of the operator/user command set."},
		{ID: "h-auth", Kind: "help", Sort: 5, Visible: true, Title: "Authentication & tokens", Meta: "Security",
			Href: "https://eos-docs.web.cern.ch/diopside/manual/using.html",
			Body: "KRB5, X.509, OIDC, SSS, and EOS tokens that delegate access without sharing credentials."},
		{ID: "h-rel", Kind: "help", Sort: 6, Visible: true, Title: "Diopside 5.5.1 notes", Meta: "Releases",
			Href: "https://eos-docs.web.cern.ch/diopside/releases/diopside-release.html#v5-5-1-diopside",
			Body: "Current stable line, 26 August 2026. Built on XRootD 5/6 with HTTP-TPC integrity work."},
	}
}

func (s *Store) ensureNewsWorkshop2027() error {
	var n int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM news WHERE id='n-ws27'`).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	return s.UpsertNews(News{
		ID: "n-ws27", Sort: 0, Visible: true,
		Title:     "Stay tuned for the next EOS workshop in spring 2027",
		DateLabel: "22 September 2026",
		Href:      "/workshops",
		Body:      "The tenth workshop is behind us. The next EOS workshop is planned for spring 2027 - dates and the Indico page will be announced here.",
	})
}

func (s *Store) ensureHelpCards() error {
	var n int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM cards WHERE kind='help'`).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	for _, c := range helpCards() {
		if err := s.UpsertCard(c); err != nil {
			return err
		}
	}
	return nil
}
