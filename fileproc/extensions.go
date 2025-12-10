// Package fileproc handles file processing, collection, and output formatting.
package fileproc

import "github.com/ivuorinen/gibidify/shared"

// getImageExtensions returns the default image file extensions.
func getImageExtensions() map[string]bool {
	return map[string]bool{
		".png":  true,
		".jpg":  true,
		".jpeg": true,
		".gif":  true,
		".bmp":  true,
		".tiff": true,
		".tif":  true,
		".svg":  true,
		".webp": true,
		".ico":  true,
	}
}

// getBinaryExtensions returns the default binary file extensions.
func getBinaryExtensions() map[string]bool {
	return map[string]bool{
		// Executables and libraries
		".exe":   true,
		".dll":   true,
		".so":    true,
		".dylib": true,
		".bin":   true,
		".o":     true,
		".a":     true,
		".lib":   true,

		// Compiled bytecode
		".jar":   true,
		".class": true,
		".pyc":   true,
		".pyo":   true,

		// Data files
		".dat":      true,
		".db":       true,
		".sqlite":   true,
		".ds_store": true,

		// Documents
		".pdf": true,

		// Archives
		".zip": true,
		".tar": true,
		".gz":  true,
		".bz2": true,
		".xz":  true,
		".7z":  true,
		".rar": true,

		// Fonts
		".ttf":   true,
		".otf":   true,
		".woff":  true,
		".woff2": true,

		// Media files
		".mp3":  true,
		".mp4":  true,
		".avi":  true,
		".mov":  true,
		".wmv":  true,
		".flv":  true,
		".webm": true,
		".ogg":  true,
		".wav":  true,
		".flac": true,
	}
}

// getLanguageMap returns the default language mappings.
func getLanguageMap() map[string]string {
	return map[string]string{
		// Systems programming
		".go":  "go",
		".c":   "c",
		".cpp": "cpp",
		".h":   "c",
		".hpp": "cpp",
		".rs":  "rust",

		// Scripting languages
		".py":  "python",
		".rb":  "ruby",
		".pl":  "perl",
		".lua": "lua",
		".php": "php",

		// Web technologies
		".js":   "javascript",
		".ts":   "typescript",
		".jsx":  "javascript",
		".tsx":  "typescript",
		".html": "html",
		".htm":  "html",
		".css":  "css",
		".scss": "scss",
		".sass": "sass",
		".less": "less",
		".vue":  "vue",

		// JVM languages
		".java":  "java",
		".scala": "scala",
		".kt":    "kotlin",
		".clj":   "clojure",

		// .NET languages
		".cs": "csharp",
		".vb": "vbnet",
		".fs": "fsharp",

		// Apple platforms
		".swift": "swift",
		".m":     "objc",
		".mm":    "objcpp",

		// Shell scripts
		".sh":   "bash",
		".bash": "bash",
		".zsh":  "zsh",
		".fish": "fish",
		".ps1":  "powershell",
		".bat":  "batch",
		".cmd":  "batch",

		// Data formats
		".json": shared.FormatJSON,
		".yaml": shared.FormatYAML,
		".yml":  shared.FormatYAML,
		".toml": "toml",
		".xml":  "xml",
		".sql":  "sql",

		// Documentation
		".md":  shared.FormatMarkdown,
		".rst": "rst",
		".tex": "latex",

		// Functional languages
		".hs":  "haskell",
		".ml":  "ocaml",
		".mli": "ocaml",
		".elm": "elm",
		".ex":  "elixir",
		".exs": "elixir",
		".erl": "erlang",
		".hrl": "erlang",

		// Other languages
		".r":    "r",
		".dart": "dart",
		".nim":  "nim",
		".nims": "nim",
	}
}
