package home

import "testing"

func TestAppendRepositoryShapes(t *testing.T) {
	entry := []string{"- name: notes", "  remote: r", "  branch: main"}
	cases := map[string]struct {
		content string
		want    string
		ok      bool
	}{
		"no list yet":          {content: "", want: "repositories:\n  - name: notes\n    remote: r\n    branch: main\n", ok: true},
		"null list":            {content: "repositories:\n", want: "repositories:\n  - name: notes\n    remote: r\n    branch: main\n", ok: true},
		"empty list comment":   {content: "repositories: [] # none yet\n", want: "repositories: # none yet\n  - name: notes\n    remote: r\n    branch: main\n", ok: true},
		"no trailing newline":  {content: "repositories:\n  - name: a\n    remote: r\n    branch: main", want: "repositories:\n  - name: a\n    remote: r\n    branch: main\n  - name: notes\n    remote: r\n    branch: main\n", ok: true},
		"flow list with items": {content: "repositories: [{name: a, remote: r, branch: main}]\n", ok: false},
		"mapping not a list":   {content: "repositories:\n  a: b\n", ok: false},
	}
	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			got, ok := appendRepository(testCase.content, entry)
			if ok != testCase.ok || (ok && got != testCase.want) {
				t.Fatalf("ok %v, got:\n%s\nwant:\n%s", ok, got, testCase.want)
			}
		})
	}
}
