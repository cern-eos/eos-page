package ingest

import "testing"

func TestParseGitShortlog(t *testing.T) {
	raw := `
  9442	Elvin Alin Sindrilaru <elvin.alin.sindrilaru@cern.ch>
  9238	Andreas Joachim Peters <andreas.joachim.peters@cern.ch>
    25	Unknown <root@vmeos03.cern.ch>
     1	okilicki <ozlem.kilickiran@cern.ch>
`
	got := ParseGitShortlog(raw)
	if len(got) != 3 {
		t.Fatalf("len %d", len(got))
	}
	if got[0].Name != "Elvin Alin Sindrilaru" || got[0].Commits != 9442 || got[0].Sort != 1 {
		t.Fatalf("first %+v", got[0])
	}
	if got[1].Name != "Andreas Joachim Peters" || got[1].Email != "andreas.joachim.peters@cern.ch" {
		t.Fatalf("second %+v", got[1])
	}
	if got[2].Name != "okilicki" {
		t.Fatalf("third %+v", got[2])
	}
}
