package store

import "strings"

func conferenceWorkshops() []Workshop {
	return []Workshop{
		{ID: "1471803", Title: "CHEP 2026", Year: 2026, Start: "2026-05-25", End: "2026-05-28", Location: "Bangkok", Kind: "conference", URL: "https://indico.cern.ch/event/1471803/", Visible: true, Sort: 100},
		{ID: "1338689", Title: "CHEP 2024", Year: 2024, Start: "2024-10-01", End: "2024-10-31", Location: "", Kind: "conference", URL: "https://indico.cern.ch/event/1338689/", Visible: true, Sort: 101},
		{ID: "995485", Title: "HEPiX Spring 2021", Year: 2021, Start: "2021-03-17", End: "2021-03-17", Location: "Online", Kind: "conference", URL: "https://indico.cern.ch/event/995485/", Visible: true, Sort: 102},
		{ID: "773049", Title: "CHEP 2019", Year: 2019, Start: "2019-11-04", End: "2019-11-08", Location: "Adelaide", Kind: "conference", URL: "https://indico.cern.ch/event/773049/", Visible: true, Sort: 103},
		{ID: "765497", Title: "HEPiX Spring 2019", Year: 2019, Start: "2019-03-27", End: "2019-03-27", Location: "", Kind: "conference", URL: "https://indico.cern.ch/event/765497/", Visible: true, Sort: 104},
		{ID: "730908", Title: "HEPiX Autumn 2018", Year: 2018, Start: "2018-10-11", End: "2018-10-11", Location: "", Kind: "conference", URL: "https://indico.cern.ch/event/730908/", Visible: true, Sort: 105},
		{ID: "587955", Title: "CHEP 2018", Year: 2018, Start: "2018-07-09", End: "2018-07-13", Location: "Sofia", Kind: "conference", URL: "https://indico.cern.ch/event/587955/", Visible: true, Sort: 106},
		{ID: "505613", Title: "CHEP 2016", Year: 2016, Start: "2016-10-10", End: "2016-10-14", Location: "San Francisco", Kind: "conference", URL: "https://indico.cern.ch/event/505613/", Visible: true, Sort: 107},
		{ID: "304944", Title: "CHEP 2015", Year: 2015, Start: "2015-04-13", End: "2015-04-17", Location: "Okinawa", Kind: "conference", URL: "https://indico.cern.ch/event/304944/", Visible: true, Sort: 108},
		{ID: "149557", Title: "CHEP 2012", Year: 2012, Start: "2012-05-21", End: "2012-05-25", Location: "", Kind: "conference", URL: "https://indico.cern.ch/event/149557/", Visible: true, Sort: 109},
		{ID: "93877", Title: "ACAT 2011", Year: 2011, Start: "2011-09-05", End: "2011-09-05", Location: "", Kind: "conference", URL: "https://indico.cern.ch/event/93877/", Visible: true, Sort: 110},
	}
}

func conferenceTalks() []Talk {
	return []Talk{
		{ID: "6967397", EventID: "1471803", Year: 2026, Title: "Real-Time I/O Traffic Shaping in EOS",
			Speakers: "Gianmaria Del Monte", Session: "28 May 2026", Start: "2026-05-28",
			URL: "https://indico.cern.ch/event/1471803/contributions/6967397/",
			Abstract: "A real-time I/O monitoring and traffic-shaping framework for EOS, to regulate throughput and improve fairness under competing workloads."},
		{ID: "6966464", EventID: "1471803", Year: 2026, Title: "Exabyte-Scale Automation, Alarms and Monitoring at CERN",
			Speakers: "Octavian-Mihai Matei", Session: "28 May 2026", Start: "2026-05-28",
			URL: "https://indico.cern.ch/event/1471803/contributions/6966464/",
			Abstract: "Running CERN EOS at very large scale: monitoring, alerting, automation, fault detection and HL-LHC preparation. More than 800 storage nodes across eight independent EOS instances."},
		{ID: "6967394", EventID: "1471803", Year: 2026, Title: "CERN Storage technology explorations: Adding NFS 4.2 as a Strategic Protocol for EOS",
			Speakers: "Andreas Joachim Peters, Elvin Alin Sindrilaru, Luca Mascetti", Session: "27 May 2026", Start: "2026-05-27",
			URL: "https://indico.cern.ch/event/1471803/contributions/6967394/",
			Abstract: "Adds NFS 4.2 as a native-access protocol for EOS, with a user-space NFS server connected through a VFS plug-in as an alternative to eosxd FUSE/XRootD on the LAN."},
		{ID: "6967329", EventID: "1471803", Year: 2026, Title: "Data Integrity and Recovery System for Distributed Storage in ALICE",
			Speakers: "Andreea Prigoreanu", Session: "27 May 2026", Start: "2026-05-27",
			URL: "https://indico.cern.ch/event/1471803/contributions/6967329/",
			Abstract: "An ALICE-wide EOS integrity and recovery system that collects FSCK reports over HTTP, analyses errors and feeds ALICE recovery procedures."},
		{ID: "6967382", EventID: "1471803", Year: 2026, Title: "Enhanced Data Integrity for Reliable WLCG Third-Party Copy Transfers",
			Speakers: "Hugo Gonzalez Labrador, Cedric Caffy, Mihai Patrascoiu", Session: "27 May 2026", Start: "2026-05-27",
			URL: "https://indico.cern.ch/event/1471803/contributions/6967382/",
			Abstract: "Stronger end-to-end integrity for HTTP Third-Party Copy, covering EOS and CTA backends and WLCG/FTS transfer workflows."},
		{ID: "6966819", EventID: "1471803", Year: 2026, Title: "Exploring AI-Assisted Coding for Storage Systems: Practical Examples and Preliminary Evaluation",
			Speakers: "Andreas Joachim Peters, Luca Mascetti", Session: "26 May 2026", Start: "2026-05-26",
			URL: "https://indico.cern.ch/event/1471803/contributions/6966819/",
			Abstract: "AI-assisted development on EOS and related CERN storage software, including I/O traffic shaping, erasure-coding performance, signature hardening, the terminal UI, XRootD authentication and FUSE behaviour."},
		{ID: "6967322", EventID: "1471803", Year: 2026, Title: "Intelligent Orchestration of Petabyte-Scale Data Staging for Physics Workflows",
			Speakers: "Alice-Florenta Suiu", Session: "25 May 2026", Start: "2026-05-25",
			URL: "https://indico.cern.ch/event/1471803/contributions/6967322/",
			Abstract: "ALICE's EOSALICEO2 high-performance disk buffer for data-taking and processing, and the aggregate read throughput it has to sustain."},
		{ID: "6967383", EventID: "1471803", Year: 2026, Title: "A new Rucio Service at CERN for Emerging and Established Experiments",
			Speakers: "Hugo Gonzalez Labrador", Session: "25 May 2026", Start: "2026-05-25",
			URL: "https://indico.cern.ch/event/1471803/contributions/6967383/",
			Abstract: "A managed CERN data-management service combining Rucio, EOS, CTA and FTS for small and medium-sized experiments."},
		{ID: "6011007", EventID: "1338689", Year: 2024, Title: "Integrating the Perlmutter HPC system in the ALICE Grid",
			Speakers: "Sergiu Weisz", Session: "October 2024", Start: "2024-10-01",
			URL: "https://indico.cern.ch/event/1338689/contributions/6011007/",
			Abstract: "NERSC Perlmutter on the ALICE Grid, including analysis workloads against an EOS instance at LBNL shared with the main Tier-2 site."},
		{ID: "4272784", EventID: "995485", Year: 2021, Title: "CTA production experience",
			Speakers: "Julien Leduc", Session: "17 March 2021", Start: "2021-03-17",
			URL: "https://indico.cern.ch/event/995485/contributions/4272784/",
			Abstract: "CERN EOSCTA production architecture: CTA as the tape backend to EOS, and the migration of the major LHC experiments."},
		{ID: "3474409", EventID: "773049", Year: 2019, Title: "EOS architectural evolution and strategic development directions",
			Speakers: "Andreas Joachim Peters", Session: "7 November 2019", Start: "2019-11-07",
			URL: "https://indico.cern.ch/event/773049/contributions/3474409/",
			Abstract: "Namespace and FUSE redesign, scalability, draining, LRU, filesystem consistency, operational reliability and the future EOS architecture."},
		{ID: "3474421", EventID: "773049", Year: 2019, Title: "Erasure Coding for production in the EOS Open Storage system",
			Speakers: "Andreas Joachim Peters, Michal Kamil Simon, Elvin Alin Sindrilaru", Session: "7 November 2019", Start: "2019-11-07",
			URL: "https://indico.cern.ch/event/773049/contributions/3474421/",
			Slides: "https://indico.cern.ch/event/773049/contributions/3474421/attachments/1940055/3217488/CHEP_2019_EC_Presentation.pdf",
			Abstract: "Deploying erasure coding in production EOS, including migration from dual replicas to EC layouts and operational tests."},
		{ID: "3474476", EventID: "773049", Year: 2019, Title: "Using the RichACL Standard for Access Control in EOS",
			Speakers: "Giuseppe Lo Presti, Rainer Toebbicke, Andreas Joachim Peters", Session: "7 November 2019", Start: "2019-11-07",
			URL: "https://indico.cern.ch/event/773049/contributions/3474476/",
			Abstract: "RichACL-style access control in EOS, including interoperability with NFSv4 and Windows/Samba ACLs."},
		{ID: "t-chep2019-declarative", EventID: "773049", Year: 2019, Title: "EOS Erasure Coding plug-in as a case study for the XRootD client declarative API",
			Speakers: "Andrew Bohdan Hanushevsky", Session: "CHEP 2019", Start: "2019-11-07",
			URL: "https://indico.cern.ch/event/773049/timetable/",
			Abstract: "The EOS erasure-coding plug-in as a practical example of the XRootD client declarative API."},
		{ID: "t-chep2019-cernbox", EventID: "773049", Year: 2019, Title: "Migration of user and project spaces with EOS/CERNBox: experience on scaling and large-scale operations",
			Speakers: "Luca Mascetti", Session: "CHEP 2019", Start: "2019-11-07",
			URL: "https://indico.cern.ch/event/773049/timetable/",
			Abstract: "Large-scale migration and operation of CERNBox user and project storage backed by EOS."},
		{ID: "3351198", EventID: "765497", Year: 2019, Title: "Storage services at CERN",
			Speakers: "Enrico Bocchi", Session: "27 March 2019", Start: "2019-03-27",
			URL: "https://indico.cern.ch/event/765497/contributions/3351198/",
			Abstract: "CERN storage infrastructure with EOS as the high-performance distributed filesystem for physics data and the backend for CERNBox."},
		{ID: "t-hepix2018-karavakis", EventID: "730908", Year: 2018, Title: "Latest developments of the CERN Data Management Tools",
			Speakers: "Edward Karavakis", Session: "11 October 2018", Start: "2018-10-11",
			URL: "https://indico.cern.ch/event/730908/",
			Abstract: "CERN IT Storage data-management technologies including EOS, DPM and FTS, and the migration of about 14,000 CERNBox users to the newer EOS architecture."},
		{ID: "3153167", EventID: "730908", Year: 2018, Title: "Storage at CERN",
			Speakers: "Cristian Contescu", Session: "11 October 2018", Start: "2018-10-11",
			URL: "https://indico.cern.ch/event/730908/contributions/3153167/",
			Abstract: "CERN storage services, with EOS as the high-performance distributed filesystem and the backend of CERNBox."},
		{ID: "t-chep2018-ecosystem", EventID: "587955", Year: 2018, Title: "EOS Open Storage - evolution of an ecosystem for scientific data repositories",
			Speakers: "Andreas Joachim Peters et al.", Session: "12 July 2018", Start: "2018-07-12",
			URL: "https://indico.cern.ch/event/587955/",
			Slides: "https://indico.cern.ch/event/587955/contributions/contributions.pdf",
			Abstract: "How EOS evolved from CERN disk-only analysis storage into a broader open scientific platform. About 250 PB across CERN data centres at the time."},
		{ID: "2230946", EventID: "505613", Year: 2016, Title: "Global EOS: exploring the 300-ms-latency region",
			Speakers: "Luca Mascetti", Session: "13 October 2016", Start: "2016-10-13",
			URL: "https://indico.cern.ch/event/505613/contributions/2230946/",
			Abstract: "A geographically distributed EOS deployment spanning CERN, Wigner, AARNet Melbourne and ASGC Taipei, including paths over 300 ms latency."},
		{ID: "1672198", EventID: "304944", Year: 2015, Title: "EOS as the present and future solution for data storage at CERN",
			Speakers: "Andreas Joachim Peters, Dirk Duellmann", Session: "16 April 2015", Start: "2015-04-16",
			URL: "https://indico.cern.ch/event/304944/contributions/1672198/",
			Abstract: "EOS at the start of LHC Run II: architecture, multi-PB deployments, dual data centres, JBOD, quality of service, disk versus tape, and the roadmap."},
		{ID: "1386169", EventID: "149557", Year: 2012, Title: "Overview of storage operations at CERN",
			Speakers: "Jan Iven, Massimo Lamanna et al.", Session: "May 2012", Start: "2012-05-21",
			URL: "https://indico.cern.ch/event/149557/contributions/1386169/",
			Abstract: "An early operational comparison of EOS and CASTOR, and the work to turn EOS into a production-quality service."},
		{ID: "2118119", EventID: "93877", Year: 2011, Title: "The EOS disk storage system at CERN",
			Speakers: "Andreas Joachim Peters", Session: "5 September 2011", Start: "2011-09-05",
			URL: "https://indico.cern.ch/event/93877/contributions/2118119/",
			Abstract: "One of the first public EOS talks after the 2010 start: disk-only storage, I/O scheduling, scalability, EOSCMS and EOSATLAS, and the early roadmap."},
	}
}

func workshopTalksCard() Card {
	return Card{
		ID: "r-search", Kind: "resource", Sort: 2, Visible: true, Title: "Workshop talks",
		Href: "/search?kind=workshop",
		Body: "Search every EOS workshop contribution since 2018: slides, recordings, and abstracts.",
	}
}

func presentationsCard() Card {
	return Card{
		ID: "r-pres", Kind: "resource", Sort: 3, Visible: true, Title: "External Presentations",
		Href: "/search?kind=external",
		Body: "A collection of conference talks about EOS outside the yearly workshop at CERN.",
	}
}

func (s *Store) EnsureExternalPresentations() error {
	for _, w := range conferenceWorkshops() {
		if err := s.UpsertWorkshop(w); err != nil {
			return err
		}
	}
	for _, t := range conferenceTalks() {
		if err := s.UpsertTalk(t); err != nil {
			return err
		}
	}
	if err := s.UpsertCard(workshopTalksCard()); err != nil {
		return err
	}
	if err := s.UpsertCard(presentationsCard()); err != nil {
		return err
	}
	if old := s.Setting("presentations_url"); old == "" || strings.Contains(strings.ToLower(old), "cernbox") || strings.Contains(old, "Kl0hxpeA5bFQ4Ho") || old == "/presentations" {
		if err := s.SetSetting("presentations_url", "/search?kind=external"); err != nil {
			return err
		}
	}
	return nil
}
