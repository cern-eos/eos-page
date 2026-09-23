package main

import (
	"bufio"
	"bytes"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func loadDotEnv() {
	for _, path := range dotenvCandidates() {
		applied, err := applyDotEnvFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			log.Printf("warning: could not read %s: %v", path, err)
			continue
		}
		if applied > 0 {
			log.Printf("loaded %s", path)
		}
		return
	}
}

func dotenvCandidates() []string {
	var out []string
	if p := strings.TrimSpace(os.Getenv("ENV_FILE")); p != "" {
		out = append(out, p)
	}
	out = append(out, ".env")
	if wd, err := os.Getwd(); err == nil {
		out = append(out, filepath.Join(wd, ".env"))
	}
	out = append(out, filepath.Join(envLookup("DATA_DIR", "data"), ".env"))
	if exe, err := os.Executable(); err == nil {
		out = append(out, filepath.Join(filepath.Dir(exe), ".env"))
	}
	seen := map[string]bool{}
	uniq := out[:0]
	for _, p := range out {
		p = filepath.Clean(p)
		if seen[p] {
			continue
		}
		seen[p] = true
		uniq = append(uniq, p)
	}
	return uniq
}

func envLookup(k, def string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		return v
	}
	return def
}

func applyDotEnvFile(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	return applyDotEnv(data), nil
}

func applyDotEnv(data []byte) int {
	data = bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))
	n := 0
	sc := bufio.NewScanner(bytes.NewReader(data))
	for sc.Scan() {
		line := strings.TrimSpace(strings.TrimSuffix(sc.Text(), "\r"))
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		if key == "" || strings.ContainsAny(key, " \t") {
			continue
		}
		val = unquoteEnv(strings.TrimSpace(val))
		if val == "" {
			continue
		}
		if cur := strings.TrimSpace(os.Getenv(key)); cur != "" {
			continue
		}
		if err := os.Setenv(key, val); err != nil {
			continue
		}
		n++
	}
	return n
}

func unquoteEnv(v string) string {
	if i := strings.Index(v, " #"); i >= 0 {
		v = strings.TrimSpace(v[:i])
	}
	if len(v) >= 2 {
		if (v[0] == '"' && v[len(v)-1] == '"') || (v[0] == '\'' && v[len(v)-1] == '\'') {
			return v[1 : len(v)-1]
		}
	}
	return v
}
