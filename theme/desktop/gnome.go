package desktop

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image/color"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	ft "fyne.io/fyne/v2/theme"
	"github.com/godbus/dbus/v5"
	"github.com/srwiley/oksvg"
)

type GnomeFlag uint8

const (
	// GnomeFlagAutoReload is a flag that indicates that the theme should be reloaded when
	// the gtk theme or icon theme changes.
	GnomeFlagAutoReload GnomeFlag = iota
)

// mapping to gnome/gtk icon names.
var gnomeIconMap = map[fyne.ThemeIconName]string{
	ft.IconNameInfo:     "dialog-information",
	ft.IconNameError:    "dialog-error",
	ft.IconNameQuestion: "dialog-question",

	ft.IconNameFolder:     "folder",
	ft.IconNameFolderNew:  "folder-new",
	ft.IconNameFolderOpen: "folder-open",
	ft.IconNameHome:       "go-home",
	ft.IconNameDownload:   "download",

	ft.IconNameDocument:        "document",
	ft.IconNameFileImage:       "image",
	ft.IconNameFileApplication: "binary",
	ft.IconNameFileText:        "text",
	ft.IconNameFileVideo:       "video",
	ft.IconNameFileAudio:       "audio",
	ft.IconNameComputer:        "computer",
	ft.IconNameMediaPhoto:      "photo",
	ft.IconNameMediaVideo:      "video",
	ft.IconNameMediaMusic:      "music",

	ft.IconNameConfirm: "dialog-apply",
	ft.IconNameCancel:  "cancel",

	ft.IconNameCheckButton:        "checkbox-symbolic",
	ft.IconNameCheckButtonChecked: "checkbox-checked-symbolic",
	ft.IconNameRadioButton:        "radio-symbolic",
	ft.IconNameRadioButtonChecked: "radio-checked-symbolic",

	ft.IconNameArrowDropDown:   "arrow-down",
	ft.IconNameArrowDropUp:     "arrow-up",
	ft.IconNameNavigateNext:    "go-right",
	ft.IconNameNavigateBack:    "go-left",
	ft.IconNameMoveDown:        "go-down",
	ft.IconNameMoveUp:          "go-up",
	ft.IconNameSettings:        "document-properties",
	ft.IconNameHistory:         "history-view",
	ft.IconNameList:            "view-list",
	ft.IconNameGrid:            "view-grid",
	ft.IconNameColorPalette:    "color-select",
	ft.IconNameColorChromatic:  "color-select",
	ft.IconNameColorAchromatic: "color-picker-grey",
}

// Map Fyne colorname to Adwaita/GTK color names
// See https://gnome.pages.gitlab.gnome.org/libadwaita/doc/main/named-colors.html
var gnomeColorMap = map[fyne.ThemeColorName]string{
	ft.ColorNameBackground:      "theme_bg_color,window_bg_color",
	ft.ColorNameForeground:      "theme_text_color,view_fg_color",
	ft.ColorNameButton:          "theme_base_color,window_bg_color",
	ft.ColorNameInputBackground: "theme_base_color,window_bg_color",
	ft.ColorNamePrimary:         "accent_color,success_color",
	ft.ColorNameError:           "error_color",
}

// Script to get the colors from the Gnome GTK/Adwaita theme.
const gjsColorScript = `
let gtkVersion = Number(ARGV[0] || 4);
imports.gi.versions.Gtk = gtkVersion + ".0";

const { Gtk, Gdk } = imports.gi;
if (gtkVersion === 3) {
  Gtk.init(null);
} else {
  Gtk.init();
}

const colors = {};
const win = new Gtk.Window();
const ctx = win.get_style_context();
const colorMap = %s;

for (let col in colorMap) {
  let [ok, bg] = [false, null];
  let found = false;
  colorMap[col].split(",").forEach((fetch) => {
    [ok, bg] = ctx.lookup_color(fetch);
    if (ok && !found) {
      found = true;
      colors[col] = [bg.red, bg.green, bg.blue, bg.alpha];
    }
  });
}

print(JSON.stringify(colors));
`

// Script to get icons from theme.
const gjsIconsScript = `
let gtkVersion = Number(ARGV[0] || 4);
imports.gi.versions.Gtk = gtkVersion + ".0";
const iconSize = 32; // can be 8, 16, 24, 32, 48, 64, 96

const { Gtk, Gdk } = imports.gi;
if (gtkVersion === 3) {
  Gtk.init(null);
} else {
  Gtk.init();
}

let iconTheme = null;
const icons = %s; // the icon list to get
const iconset = {};

if (gtkVersion === 3) {
  iconTheme = Gtk.IconTheme.get_default();
} else {
  iconTheme = Gtk.IconTheme.get_for_display(Gdk.Display.get_default());
}

icons.forEach((name) => {
  try {
    if (gtkVersion === 3) {
      const icon = iconTheme.lookup_icon(name, iconSize, 0);
      iconset[name] = icon.get_filename();
    } else {
      const icon = iconTheme.lookup_icon(name, null, null, iconSize, null, 0);
      iconset[name] = icon.file.get_path();
    }
  } catch (e) {
    iconset[name] = null;
  }
});

print(JSON.stringify(iconset));
`

// GnomeTheme theme, based on the Gnome desktop manager. This theme uses GJS and gsettings to get
// the colors and font from the Gnome desktop.
type GnomeTheme struct {
	colors map[fyne.ThemeColorName]color.Color
	icons  map[string]string

	fontScaleFactor float32
	font            fyne.Resource
	fontSize        float32
	variant         *fyne.ThemeVariant
	iconCache       map[string]fyne.Resource
}

// Color returns the color for the given color name
//
// Implements: fyne.Theme
func (gnome *GnomeTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if col, ok := gnome.colors[name]; ok {
		return col
	}

	if gnome.variant == nil {
		return ft.DefaultTheme().Color(name, *gnome.variant)
	}

	return ft.DefaultTheme().Color(name, variant)
}

// Font returns the font for the given name.
//
// Implements: fyne.Theme
func (gnome *GnomeTheme) Font(s fyne.TextStyle) fyne.Resource {
	if gnome.font == nil {
		return ft.DefaultTheme().Font(s)
	}
	return gnome.font
}

// Icon returns the icon for the given name.
//
// Implements: fyne.Theme
func (gnome *GnomeTheme) Icon(i fyne.ThemeIconName) fyne.Resource {

	if icon, found := gnomeIconMap[i]; found {
		if resource := gnome.loadIcon(icon); resource != nil {
			return resource
		}
	}
	//log.Println(i)
	return ft.DefaultTheme().Icon(i)
}

// Size returns the size for the given name. It will scale the detected Gnome font size
// by the Gnome font factor.
//
// Implements: fyne.Theme
func (g *GnomeTheme) Size(s fyne.ThemeSizeName) float32 {
	switch s {
	case ft.SizeNameText:
		return g.fontScaleFactor * g.fontSize
	}
	return ft.DefaultTheme().Size(s)
}

// applyColors sets the colors for the Gnome theme. Colors are defined by a GJS script.
func (gnome *GnomeTheme) applyColors(gtkVersion int, wg *sync.WaitGroup) {

	if wg != nil {
		defer wg.Done()
	}
	// we will call gjs to get the colors
	gjs, err := exec.LookPath("gjs")
	if err != nil {
		log.Println("To activate the theme, please install gjs", err)
		return
	}

	// create a temp file to store the colors
	f, err := ioutil.TempFile("", "fyne-theme-gnome-*.js")
	if err != nil {
		log.Println(err)
		return
	}
	defer os.Remove(f.Name())

	// generate the js object from gnomeColorMap
	colormap := "{\n"
	for col, fetch := range gnomeColorMap {
		colormap += fmt.Sprintf(`    "%s": "%s",`+"\n", col, fetch)
	}
	colormap += "}"

	// write the script to the temp file
	script := fmt.Sprintf(gjsColorScript, colormap)
	_, err = f.WriteString(script)
	if err != nil {
		log.Println(err)
		return
	}

	// run the script
	cmd := exec.Command(gjs,
		f.Name(), strconv.Itoa(gtkVersion),
		fmt.Sprintf("%0.2f", 1.0),
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Println("gjs error:", err, string(out))
		return
	}

	// decode json
	var colors map[fyne.ThemeColorName][]float32
	err = json.Unmarshal(out, &colors)
	if err != nil {
		log.Println("gjs error:", err, string(out))
		return
	}
	for name, rgba := range colors {
		// convert string arry to colors
		gnome.colors[name] = gnome.parseColor(rgba)
	}

	gnome.calculateVariant()

}

// applyFont gets the font name from gsettings and set the font size. This also calls
// setFont() to set the font.
func (gnome *GnomeTheme) applyFont(wg *sync.WaitGroup) {

	if wg != nil {
		defer wg.Done()
	}

	gnome.font = ft.TextFont()
	// call gsettings get org.gnome.desktop.interface font-name
	cmd := exec.Command("gsettings", "get", "org.gnome.desktop.interface", "font-name")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Println(err)
		log.Println(string(out))
		return
	}
	// try to get the font as a TTF file
	fontFile := strings.TrimSpace(string(out))
	fontFile = strings.Trim(fontFile, "'")
	// the fontFile string is in the format: Name size, eg: "Sans Bold 12", so get the size
	parts := strings.Split(fontFile, " ")
	fontSize := parts[len(parts)-1]
	// convert the size to a float
	size, err := strconv.ParseFloat(fontSize, 32)
	// apply this to the fontScaleFactor
	gnome.fontSize = float32(size)

	// try to get the font as a TTF file
	gnome.setFont(strings.Join(parts[:len(parts)-1], " "))
}

// applyFontScale find the font scaling factor in settings.
func (gnome *GnomeTheme) applyFontScale(wg *sync.WaitGroup) {
	if wg != nil {
		defer wg.Done()
	}
	// for any error below, we will use the default
	gnome.fontScaleFactor = 1

	// call gsettings get org.gnome.desktop.interface text-scaling-factor
	cmd := exec.Command("gsettings", "get", "org.gnome.desktop.interface", "text-scaling-factor")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return
	}

	// get the text scaling factor
	ts := strings.TrimSpace(string(out))
	textScale, err := strconv.ParseFloat(ts, 32)
	if err != nil {
		return
	}

	// return the text scaling factor
	gnome.fontScaleFactor = float32(textScale)
}

// applyIcons gets the icon theme from gsettings and call GJS script to get the icon set.
func (gnome *GnomeTheme) applyIcons(gtkVersion int, wg *sync.WaitGroup) {

	if wg != nil {
		defer wg.Done()
	}

	gjs, err := exec.LookPath("gjs")
	if err != nil {
		log.Println("To activate the theme, please install gjs", err)
		return
	}
	// create the list of icon to get
	var icons []string
	for _, icon := range gnomeIconMap {
		icons = append(icons, icon)
	}
	iconSet := "[\n"
	for _, icon := range icons {
		iconSet += fmt.Sprintf(`    "%s",`+"\n", icon)
	}
	iconSet += "]"

	gjsIconList := fmt.Sprintf(gjsIconsScript, iconSet)

	// write the script to a temp file
	f, err := ioutil.TempFile("", "fyne-theme-gnome-*.js")
	if err != nil {
		log.Println(err)
		return
	}
	defer os.Remove(f.Name())

	// write the script to the temp file
	_, err = f.WriteString(gjsIconList)
	if err != nil {
		log.Println(err)
		return
	}

	// Call gjs with 2 version, 3 and 4 to complete the icon, this because
	// gtk version is sometimes not available or icon is not fully completed...
	// It's a bit tricky but it works.
	for _, gtkVersion := range []string{"3", "4"} {
		// run the script
		cmd := exec.Command(gjs,
			f.Name(), gtkVersion,
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Println("gjs error:", err, string(out))
			return
		}

		tmpicons := map[string]*string{}
		// decode json to apply to the gnome theme
		err = json.Unmarshal(out, &tmpicons)
		if err != nil {
			log.Println(err)
			return
		}
		for k, v := range tmpicons {
			if _, ok := gnome.icons[k]; !ok {
				if v != nil && *v != "" {
					gnome.icons[k] = *v
				}
			}
		}
	}
}

// calculateVariant calculates the variant of the theme using the background color.
func (gnome *GnomeTheme) calculateVariant() {
	r, g, b, _ := gnome.Color(ft.ColorNameBackground, 0).RGBA()

	brightness := (r/255*299 + g/255*587 + b/255*114) / 1000
	gnome.variant = new(fyne.ThemeVariant)
	if brightness > 125 {
		*gnome.variant = ft.VariantLight
	} else {
		*gnome.variant = ft.VariantDark
	}
}

// findThemeInformation decodes the theme from the gsettings and Gtk API.
func (gnome *GnomeTheme) findThemeInformation(gtkVersion int, variant fyne.ThemeVariant) {
	// make things faster in concurrent mode
	wg := sync.WaitGroup{}
	wg.Add(4)
	go gnome.applyColors(gtkVersion, &wg)
	go gnome.applyIcons(gtkVersion, &wg)
	go gnome.applyFont(&wg)
	go gnome.applyFontScale(&wg)
	wg.Wait()
}

// getGTKVersion gets the available GTK version for the given theme. If the version cannot be
// determine, it will return 3 wich is the most common used version.
func (gnome *GnomeTheme) getGTKVersion() int {

	themename := gnome.getThemeName()
	if themename == "" {
		return 3 // default to 3
	}

	// ok so now, find if the theme is gtk4, either fallback to gtk3
	home, err := os.UserHomeDir()
	if err != nil {
		log.Println(err)
		return 3 // default to Gtk 3
	}

	possiblePaths := []string{
		home + "/.local/share/themes/",
		home + "/.themes/",
		`/usr/local/share/themes/`,
		`/usr/share/themes/`,
	}

	for _, path := range possiblePaths {
		path = filepath.Join(path, themename)
		if _, err := os.Stat(path); err == nil {
			// now check if it is gtk4 compatible
			if _, err := os.Stat(path + "gtk-4.0/gtk.css"); err == nil {
				// it is gtk4
				return 4
			}
			if _, err := os.Stat(path + "gtk-3.0/gtk.css"); err == nil {
				return 3
			}
		}
	}
	return 3 // default, but that may be a false positive now
}

// getIconThemeName return the current icon theme name.
func (gnome *GnomeTheme) getIconThemeName() string {
	// call gsettings get org.gnome.desktop.interface icon-theme
	cmd := exec.Command("gsettings", "get", "org.gnome.desktop.interface", "icon-theme")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Println(err)
		log.Println(string(out))
		return ""
	}
	themename := strings.TrimSpace(string(out))
	themename = strings.Trim(themename, "'")
	return themename
}

// getThemeName gets the current theme name.
func (gnome *GnomeTheme) getThemeName() string {
	// call gsettings get org.gnome.desktop.interface gtk-theme
	cmd := exec.Command("gsettings", "get", "org.gnome.desktop.interface", "gtk-theme")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Println(err)
		log.Println(string(out))
		return ""
	}
	themename := strings.TrimSpace(string(out))
	themename = strings.Trim(themename, "'")
	return themename
}

// loadIcon loads the icon from gnome theme, if the icon was already loaded, so the cached version is returned.
func (gnome *GnomeTheme) loadIcon(name string) (resource fyne.Resource) {
	var ok bool

	if resource, ok = gnome.iconCache[name]; ok {
		return
	}

	defer func() {
		// whatever the result is, cache it
		// even if it is nil
		gnome.iconCache[name] = resource
	}()

	if filename, ok := gnome.icons[name]; ok {
		content, err := ioutil.ReadFile(filename)
		if err != nil {
			return
		}
		if strings.HasSuffix(filename, ".svg") {
			// we need to ensure that the svg can be opened by Fyne
			buff := bytes.NewBuffer(content)
			_, err := oksvg.ReadIconStream(buff)
			if err != nil {
				// try to convert it to png with imageMagik
				log.Println("Cannot load file", filename, err, "try to convert")
				resource, err = convertSVGtoPNG(filename)
				if err != nil {
					log.Println("Cannot convert file", filename, err)
					return
				}
				return
			}
		}
		resource = fyne.NewStaticResource(filename, content)
		return
	}
	return
}

// parseColor converts a float32 array to color.Color.
func (*GnomeTheme) parseColor(col []float32) color.Color {
	return color.RGBA{
		R: uint8(col[0] * 255),
		G: uint8(col[1] * 255),
		B: uint8(col[2] * 255),
		A: uint8(col[3] * 255),
	}
}

// setFont sets the font for the theme - this method calls getFontPath() and converToTTF
// if needed.
func (gnome *GnomeTheme) setFont(fontname string) {

	fontpath, err := getFontPath(fontname)
	if err != nil {
		log.Println(err)
		return
	}

	ext := filepath.Ext(fontpath)
	if ext != ".ttf" {
		font, err := converToTTF(fontpath)
		if err != nil {
			log.Println(err)
			return
		}
		gnome.font = fyne.NewStaticResource(fontpath, font)
	} else {
		font, err := ioutil.ReadFile(fontpath)
		if err != nil {
			log.Println(err)
			return
		}
		gnome.font = fyne.NewStaticResource(fontpath, font)
	}
}

// NewGnomeTheme returns a new Gnome theme based on the given gtk version. If gtkVersion is <= 0,
// the theme will try to determine the higher Gtk version available for the current GtkTheme.
func NewGnomeTheme(gtkVersion int, flags ...GnomeFlag) fyne.Theme {
	gnome := &GnomeTheme{
		fontSize:  ft.DefaultTheme().Size(ft.SizeNameText),
		iconCache: map[string]fyne.Resource{},
		icons:     map[string]string{},
		colors:    map[fyne.ThemeColorName]color.Color{},
	}

	if gtkVersion <= 0 {
		// detect gtkVersion
		gtkVersion = gnome.getGTKVersion()
	}
	gnome.findThemeInformation(gtkVersion, ft.VariantDark)

	for _, flag := range flags {
		switch flag {
		case GnomeFlagAutoReload:
			go func() {
				// connect to setting changes to not reload the theme if the new selected is
				// not a gnome theme
				settingChan := make(chan fyne.Settings)
				fyne.CurrentApp().Settings().AddChangeListener(settingChan)

				// connect to dbus to detect theme/icon them changes
				conn, err := dbus.SessionBus()
				if err != nil {
					log.Println(err)
					return
				}
				if err := conn.AddMatchSignal(
					dbus.WithMatchObjectPath("/org/freedesktop/portal/desktop"),
					dbus.WithMatchInterface("org.freedesktop.portal.Settings"),
					dbus.WithMatchMember("SettingChanged"),
				); err != nil {
					log.Println(err)
					return
				}
				defer conn.Close()
				dbusChan := make(chan *dbus.Signal, 5)
				conn.Signal(dbusChan)

				for {
					select {
					case sig := <-dbusChan:
						// break if the current theme is not typed as GnomeTheme
						currentTheme := fyne.CurrentApp().Settings().Theme()
						if _, ok := currentTheme.(*GnomeTheme); !ok {
							return
						}
						// reload the theme if the changed setting is the Gtk theme
						for _, v := range sig.Body {
							switch v {
							case "gtk-theme", "icon-theme", "text-scaling-factor", "font-name":
								go fyne.CurrentApp().Settings().SetTheme(NewGnomeTheme(gtkVersion, flags...))
								return
							}
						}
					case s := <-settingChan:
						// leave the loop if the new theme is not a Gnome theme
						if _, isGnome := s.Theme().(*GnomeTheme); !isGnome {
							return
						}
					}
				}
			}()
		}
	}
	return gnome
}
