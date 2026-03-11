package validation

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// hexAlphaColorPattern matches 4- or 8-digit hex colors with alpha.
var hexAlphaColorPattern = regexp.MustCompile(`^#([0-9a-fA-F]{4}|[0-9a-fA-F]{8})$`)

// rgbPattern matches rgb(R, G, B) with integer values.
var rgbPattern = regexp.MustCompile(`^rgb\(\s*(\d{1,3})\s*,\s*(\d{1,3})\s*,\s*(\d{1,3})\s*\)$`)

// hslPattern matches hsl(H, S%, L%) with integer values.
var hslPattern = regexp.MustCompile(`^hsl\(\s*(\d{1,3})\s*,\s*(\d{1,3})%\s*,\s*(\d{1,3})%\s*\)$`)

// HexAlphaColor validates that value is a valid hex color with alpha (#RGBA or #RRGGBBAA).
func HexAlphaColor(field, value string) error {
	if !hexAlphaColorPattern.MatchString(value) {
		return fmt.Errorf("%s must be a valid hex color with alpha (e.g. #fffa or #a1b2c3ff)", field)
	}
	return nil
}

// RGBColor validates that value is a valid rgb(R, G, B) color string with
// each component in the range 0-255.
func RGBColor(field, value string) error {
	m := rgbPattern.FindStringSubmatch(value)
	if m == nil {
		return fmt.Errorf("%s must be a valid RGB color (e.g. rgb(255, 128, 0))", field)
	}
	for _, s := range m[1:] {
		n, _ := strconv.Atoi(s)
		if n > 255 {
			return fmt.Errorf("%s must be a valid RGB color (e.g. rgb(255, 128, 0))", field)
		}
	}
	return nil
}

// HSLColor validates that value is a valid hsl(H, S%, L%) color string with
// hue in 0-360 and saturation/lightness in 0-100.
func HSLColor(field, value string) error {
	m := hslPattern.FindStringSubmatch(value)
	if m == nil {
		return fmt.Errorf("%s must be a valid HSL color (e.g. hsl(120, 50%%, 75%%))", field)
	}
	h, _ := strconv.Atoi(m[1])
	s, _ := strconv.Atoi(m[2])
	l, _ := strconv.Atoi(m[3])
	if h > 360 || s > 100 || l > 100 {
		return fmt.Errorf("%s must be a valid HSL color (e.g. hsl(120, 50%%, 75%%))", field)
	}
	return nil
}

// CSSColor validates that value is any recognized CSS color format: hex (#RGB,
// #RRGGBB), rgb(), hsl(), or a named CSS color.
func CSSColor(field, value string) error {
	if hexColorRe.MatchString(value) {
		return nil
	}
	if rgbPattern.MatchString(value) {
		return RGBColor(field, value)
	}
	if hslPattern.MatchString(value) {
		return HSLColor(field, value)
	}
	if _, ok := cssNamedColors[strings.ToLower(value)]; ok {
		return nil
	}
	return fmt.Errorf("%s must be a valid CSS color", field)
}

// cssNamedColors contains the 148 standard CSS named colors.
var cssNamedColors = map[string]struct{}{
	"aliceblue": {}, "antiquewhite": {}, "aqua": {}, "aquamarine": {},
	"azure": {}, "beige": {}, "bisque": {}, "black": {},
	"blanchedalmond": {}, "blue": {}, "blueviolet": {}, "brown": {},
	"burlywood": {}, "cadetblue": {}, "chartreuse": {}, "chocolate": {},
	"coral": {}, "cornflowerblue": {}, "cornsilk": {}, "crimson": {},
	"cyan": {}, "darkblue": {}, "darkcyan": {}, "darkgoldenrod": {},
	"darkgray": {}, "darkgreen": {}, "darkgrey": {}, "darkkhaki": {},
	"darkmagenta": {}, "darkolivegreen": {}, "darkorange": {}, "darkorchid": {},
	"darkred": {}, "darksalmon": {}, "darkseagreen": {}, "darkslateblue": {},
	"darkslategray": {}, "darkslategrey": {}, "darkturquoise": {}, "darkviolet": {},
	"deeppink": {}, "deepskyblue": {}, "dimgray": {}, "dimgrey": {},
	"dodgerblue": {}, "firebrick": {}, "floralwhite": {}, "forestgreen": {},
	"fuchsia": {}, "gainsboro": {}, "ghostwhite": {}, "gold": {},
	"goldenrod": {}, "gray": {}, "green": {}, "greenyellow": {},
	"grey": {}, "honeydew": {}, "hotpink": {}, "indianred": {},
	"indigo": {}, "ivory": {}, "khaki": {}, "lavender": {},
	"lavenderblush": {}, "lawngreen": {}, "lemonchiffon": {}, "lightblue": {},
	"lightcoral": {}, "lightcyan": {}, "lightgoldenrodyellow": {}, "lightgray": {},
	"lightgreen": {}, "lightgrey": {}, "lightpink": {}, "lightsalmon": {},
	"lightseagreen": {}, "lightskyblue": {}, "lightslategray": {}, "lightslategrey": {},
	"lightsteelblue": {}, "lightyellow": {}, "lime": {}, "limegreen": {},
	"linen": {}, "magenta": {}, "maroon": {}, "mediumaquamarine": {},
	"mediumblue": {}, "mediumorchid": {}, "mediumpurple": {}, "mediumseagreen": {},
	"mediumslateblue": {}, "mediumspringgreen": {}, "mediumturquoise": {}, "mediumvioletred": {},
	"midnightblue": {}, "mintcream": {}, "mistyrose": {}, "moccasin": {},
	"navajowhite": {}, "navy": {}, "oldlace": {}, "olive": {},
	"olivedrab": {}, "orange": {}, "orangered": {}, "orchid": {},
	"palegoldenrod": {}, "palegreen": {}, "paleturquoise": {}, "palevioletred": {},
	"papayawhip": {}, "peachpuff": {}, "peru": {}, "pink": {},
	"plum": {}, "powderblue": {}, "purple": {}, "rebeccapurple": {},
	"red": {}, "rosybrown": {}, "royalblue": {}, "saddlebrown": {},
	"salmon": {}, "sandybrown": {}, "seagreen": {}, "seashell": {},
	"sienna": {}, "silver": {}, "skyblue": {}, "slateblue": {},
	"slategray": {}, "slategrey": {}, "snow": {}, "springgreen": {},
	"steelblue": {}, "tan": {}, "teal": {}, "thistle": {},
	"tomato": {}, "turquoise": {}, "violet": {}, "wheat": {},
	"white": {}, "whitesmoke": {}, "yellow": {}, "yellowgreen": {},
	"transparent": {},
}
