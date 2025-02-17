# go-rofi-bookmarks

A rofi plugin to search and manage browser bookmarks.

## Configuration

The application will automatically try to find your browser profiles in common locations. However, if you need to specify custom profile paths, you can create a configuration file.

### Default Configuration File Location

Create your configuration file in one of these locations:
- `~/.config/go-rofi-bookmarks/config.json`
- `$XDG_CONFIG_HOME/go-rofi-bookmarks/config.json`

### Example Configuration

An example configuration file is provided in `config.example.json`. Here's how to configure each browser:

```json
{
  "firefox": {
    "profile_path": "~/.mozilla/firefox/xxxxxxxx.default-release"
  },
  "chrome": {
    "profile_path": "~/.config/google-chrome/Default"
  }
}
```

### Finding Your Profile Path

#### Firefox
1. Open Firefox and navigate to `about:profiles`
2. Look for the profile marked as "Default"
3. Copy the "Root Directory" path

#### Chrome
The default profile is typically located at:
- Linux: `~/.config/google-chrome/Default`
- Flatpak: `~/.var/app/com.google.Chrome/config/google-chrome/Default`

Note: The application will attempt to automatically find these paths, so manual configuration is only needed if the automatic detection fails or if you want to use a different profile.
