package paths

import (
	"os"
	"path/filepath"
)

// XDG Base Directory paths
var (
	ConfigDir   string
	DataDir     string
	StateDir    string
	CacheDir    string
	PicturesDir string
	VideosDir   string

	// Heimdall-specific directories
	HeimdallConfigDir string
	HeimdallDataDir   string
	HeimdallStateDir  string
	HeimdallCacheDir  string

	// Specific paths
	UserConfigPath         string
	SchemeStatePath        string
	WallpaperPath          string
	WallpaperLinkPath      string
	WallpaperThumbnailPath string
	RecordingPath          string
	RecordingNotifPath     string

	// Directories
	TemplatesDir        string
	UserTemplatesDir    string
	ThemeDir            string
	SchemeDataDir       string
	SchemeCacheDir      string
	UserSchemeDir       string
	WallpapersDir       string
	WallpapersCacheDir  string
	ScreenshotsDir      string
	ScreenshotsCacheDir string
	RecordingsDir       string
)

func init() {
	// Initialize XDG base directories
	home, err := os.UserHomeDir()
	if err != nil {
		panic("Failed to get home directory: " + err.Error())
	}

	ConfigDir = getEnvOrDefault("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	DataDir = getEnvOrDefault("XDG_DATA_HOME", filepath.Join(home, ".local", "share"))
	StateDir = getEnvOrDefault("XDG_STATE_HOME", filepath.Join(home, ".local", "state"))
	CacheDir = getEnvOrDefault("XDG_CACHE_HOME", filepath.Join(home, ".cache"))
	PicturesDir = getEnvOrDefault("XDG_PICTURES_DIR", filepath.Join(home, "Pictures"))
	VideosDir = getEnvOrDefault("XDG_VIDEOS_DIR", filepath.Join(home, "Videos"))

	// Initialize Heimdall-specific directories
	HeimdallConfigDir = filepath.Join(ConfigDir, "heimdall")
	HeimdallDataDir = filepath.Join(DataDir, "heimdall")
	HeimdallStateDir = filepath.Join(StateDir, "heimdall")
	HeimdallCacheDir = filepath.Join(CacheDir, "heimdall")

	// Initialize specific paths
	UserConfigPath = filepath.Join(HeimdallConfigDir, "config.yaml")
	SchemeStatePath = filepath.Join(HeimdallStateDir, "scheme.json")
	WallpaperPath = filepath.Join(HeimdallStateDir, "wallpaper", "path.txt")
	WallpaperLinkPath = filepath.Join(HeimdallStateDir, "wallpaper", "current")
	WallpaperThumbnailPath = filepath.Join(HeimdallStateDir, "wallpaper", "thumbnail.jpg")
	RecordingPath = filepath.Join(HeimdallStateDir, "record", "recording.mp4")
	RecordingNotifPath = filepath.Join(HeimdallStateDir, "record", "notifid.txt")

	// Initialize directories
	TemplatesDir = filepath.Join(HeimdallDataDir, "templates")
	UserTemplatesDir = filepath.Join(HeimdallConfigDir, "templates")
	ThemeDir = filepath.Join(HeimdallStateDir, "theme")
	SchemeDataDir = filepath.Join(HeimdallDataDir, "schemes")
	SchemeCacheDir = filepath.Join(HeimdallCacheDir, "schemes")
	UserSchemeDir = filepath.Join(HeimdallConfigDir, "schemes")
	WallpapersDir = filepath.Join(PicturesDir, "Wallpapers")
	WallpapersCacheDir = filepath.Join(HeimdallCacheDir, "wallpapers")
	ScreenshotsDir = filepath.Join(PicturesDir, "Screenshots")
	ScreenshotsCacheDir = filepath.Join(HeimdallCacheDir, "screenshots")
	RecordingsDir = filepath.Join(VideosDir, "Recordings")
}

// getEnvOrDefault returns the value of an environment variable or a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// EnsureDir creates a directory if it doesn't exist
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// EnsureParentDir creates the parent directory of a path if it doesn't exist
func EnsureParentDir(path string) error {
	parent := filepath.Dir(path)
	return EnsureDir(parent)
}

// Exists checks if a path exists
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// IsDir checks if a path is a directory
func IsDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// IsFile checks if a path is a regular file
func IsFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.Mode().IsRegular()
}
