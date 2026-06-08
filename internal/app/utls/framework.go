package utls

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

type FrameworkPreset struct {
	Framework       string   `json:"framework"`
	BuildCommand    []string `json:"buildCommand"`
	OutputDirectory string   `json:"outputDirectory"`
	RunCommand      string   `json:"runCommand"`
	Port            int      `json:"port"`
	Version         string   `json:"version"`
}

var Presets = map[string]FrameworkPreset{
	"Next.js":      {Framework: "Next.js", BuildCommand: []string{"npm install", "npm run build"}, OutputDirectory: ".next", RunCommand: "npm start", Port: 3000, Version: "22"},
	"React":        {Framework: "React", BuildCommand: []string{"npm install", "npm run build"}, OutputDirectory: "build", RunCommand: "npm start", Port: 3000, Version: "22"},
	"Vue.js":       {Framework: "Vue.js", BuildCommand: []string{"npm install", "npm run build"}, OutputDirectory: "dist", RunCommand: "npm start", Port: 8080, Version: "22"},
	"Nuxt.js":      {Framework: "Nuxt.js", BuildCommand: []string{"npm install", "npm run build"}, OutputDirectory: ".output", RunCommand: "npm start", Port: 3000, Version: "22"},
	"Angular":      {Framework: "Angular", BuildCommand: []string{"npm install", "npm run build"}, OutputDirectory: "dist", RunCommand: "npm start", Port: 4200, Version: "22"},
	"SvelteKit":    {Framework: "SvelteKit", BuildCommand: []string{"npm install", "npm run build"}, OutputDirectory: "build", RunCommand: "node build", Port: 3000, Version: "22"},
	"Svelte":       {Framework: "Svelte", BuildCommand: []string{"npm install", "npm run build"}, OutputDirectory: "public", RunCommand: "npm start", Port: 5000, Version: "22"},
	"Vite":         {Framework: "Vite", BuildCommand: []string{"npm install", "npm run build"}, OutputDirectory: "dist", RunCommand: "npm start", Port: 5173, Version: "22"},
	"Astro":        {Framework: "Astro", BuildCommand: []string{"npm install", "npm run build"}, OutputDirectory: "dist", RunCommand: "npm start", Port: 4321, Version: "22"},
	"Gatsby":       {Framework: "Gatsby", BuildCommand: []string{"npm install", "npm run build"}, OutputDirectory: "public", RunCommand: "npm start", Port: 8000, Version: "22"},
	"Remix":        {Framework: "Remix", BuildCommand: []string{"npm install", "npm run build"}, OutputDirectory: "build", RunCommand: "npm start", Port: 3000, Version: "22"},
	"Express.js":   {Framework: "Express.js", BuildCommand: []string{"npm install"}, OutputDirectory: ".", RunCommand: "npm start", Port: 3000, Version: "22"},
	"Nest.js":      {Framework: "Nest.js", BuildCommand: []string{"npm install", "npm run build"}, OutputDirectory: "dist", RunCommand: "npm run start:prod", Port: 3000, Version: "22"},
	"Fastify":      {Framework: "Fastify", BuildCommand: []string{"npm install"}, OutputDirectory: ".", RunCommand: "npm start", Port: 3000, Version: "22"},
	"Bun":          {Framework: "Bun", BuildCommand: []string{"bun install"}, OutputDirectory: ".", RunCommand: "bun run start", Port: 3000, Version: "latest"},
	"FastAPI":      {Framework: "FastAPI", BuildCommand: []string{"pip install --no-cache-dir -r requirements.txt"}, OutputDirectory: ".", RunCommand: "uvicorn main:app --host 0.0.0.0 --port 8000", Port: 8000, Version: "3.12"},
	"Django":       {Framework: "Django", BuildCommand: []string{"pip install --no-cache-dir -r requirements.txt"}, OutputDirectory: ".", RunCommand: "gunicorn wsgi:application -b 0.0.0.0:8000", Port: 8000, Version: "3.12"},
	"Flask":        {Framework: "Flask", BuildCommand: []string{"pip install --no-cache-dir -r requirements.txt"}, OutputDirectory: ".", RunCommand: "gunicorn app:app -b 0.0.0.0:5000", Port: 5000, Version: "3.12"},
	"Python":       {Framework: "Python", BuildCommand: []string{"pip install --no-cache-dir -r requirements.txt"}, OutputDirectory: ".", RunCommand: "python app.py", Port: 8000, Version: "3.12"},
	"Go":           {Framework: "Go", BuildCommand: []string{"go build -o main ."}, OutputDirectory: ".", RunCommand: "./main", Port: 8080, Version: "1.22"},
	"Rust":         {Framework: "Rust", BuildCommand: []string{"cargo build --release"}, OutputDirectory: ".", RunCommand: "./app", Port: 8080, Version: "1"},
	"html-css-js":  {Framework: "html-css-js", BuildCommand: []string{}, OutputDirectory: ".", RunCommand: "serve -s . -l 80", Port: 80, Version: "latest"},
	"Other":        {Framework: "Other", BuildCommand: []string{}, OutputDirectory: ".", RunCommand: "", Port: 3000, Version: "latest"},
}

var FrameworkNames = []string{
	"Next.js", "React", "Vue.js", "Nuxt.js", "Angular", "SvelteKit", "Svelte", "Vite", "Astro",
	"Gatsby", "Remix", "Express.js", "Nest.js", "Fastify", "Bun", "FastAPI", "Django", "Flask",
	"Python", "Go", "Rust", "html-css-js", "Other",
}

func PresetFor(framework string) FrameworkPreset {
	if preset, ok := Presets[framework]; ok {
		return preset
	}
	return Presets["Other"]
}

func fileExists(dir, file string) bool {
	_, err := os.Stat(filepath.Join(dir, file))
	return err == nil
}

type packageJSON struct {
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

func detectNodeFramework(dir string) (FrameworkPreset, bool) {
	pkgPath := filepath.Join(dir, "package.json")
	data, err := os.ReadFile(pkgPath)
	if err != nil {
		return FrameworkPreset{}, false
	}
	var pkg packageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return FrameworkPreset{}, false
	}

	has := func(name string) bool {
		_, ok1 := pkg.Dependencies[name]
		_, ok2 := pkg.DevDependencies[name]
		return ok1 || ok2
	}

	if has("next") {
		return PresetFor("Next.js"), true
	}
	if has("nuxt") || has("nuxt3") {
		return PresetFor("Nuxt.js"), true
	}
	if has("@angular/core") {
		return PresetFor("Angular"), true
	}
	if has("@sveltejs/kit") {
		return PresetFor("SvelteKit"), true
	}
	if has("@remix-run/react") {
		return PresetFor("Remix"), true
	}
	if has("astro") {
		return PresetFor("Astro"), true
	}
	if has("gatsby") {
		return PresetFor("Gatsby"), true
	}
	if has("svelte") {
		return PresetFor("Svelte"), true
	}
	if has("@nestjs/core") {
		return PresetFor("Nest.js"), true
	}
	if has("express") {
		return PresetFor("Express.js"), true
	}
	if has("fastify") {
		return PresetFor("Fastify"), true
	}
	if has("react") && (has("react-scripts") || has("react-dom")) {
		if has("vite") {
			return PresetFor("Vite"), true
		}
		return PresetFor("React"), true
	}
	if has("vue") {
		return PresetFor("Vue.js"), true
	}
	if has("vite") {
		return PresetFor("Vite"), true
	}
	if fileExists(dir, "bun.lockb") || fileExists(dir, "bun.lock") {
		return PresetFor("Bun"), true
	}
	return PresetFor("Express.js"), true // Generic node fallback
}

func DetectFramework(dir string) FrameworkPreset {
	if fileExists(dir, "package.json") {
		if preset, ok := detectNodeFramework(dir); ok {
			return preset
		}
	}

	if fileExists(dir, "requirements.txt") || fileExists(dir, "pyproject.toml") {
		reqPath := filepath.Join(dir, "requirements.txt")
		data, err := os.ReadFile(reqPath)
		reqs := ""
		if err == nil {
			reqs = strings.ToLower(string(data))
		}
		if strings.Contains(reqs, "django") || fileExists(dir, "manage.py") {
			return PresetFor("Django")
		}
		if strings.Contains(reqs, "fastapi") {
			return PresetFor("FastAPI")
		}
		if strings.Contains(reqs, "flask") {
			return PresetFor("Flask")
		}
		return PresetFor("Python")
	}

	if fileExists(dir, "go.mod") {
		return PresetFor("Go")
	}
	if fileExists(dir, "Cargo.toml") {
		return PresetFor("Rust")
	}
	if fileExists(dir, "index.html") {
		return PresetFor("html-css-js")
	}

	return PresetFor("Other")
}
