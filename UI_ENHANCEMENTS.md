# QUIBIT CLI - Enhanced UI/UX Documentation

## 🎨 Perubahan yang Telah Dilakukan

### 1. **3D Title dengan Efek Shadow** ✅

- Implementasi 3D extrusion effect seperti HEXSEC GPT
- Shadow menggunakan deep purple untuk kedalaman visual
- Adaptive berdasarkan terminal width:
  - **120+ columns**: Full 3D effect dengan shadow besar (dx=2, dy=1)
  - **70-120 columns**: Medium 3D effect (dx=1, dy=1)
  - **55-70 columns**: Standard colored letters
  - **30-55 columns**: Compact version
  - **<30 columns**: Ultra-compact fallback

### 2. **Tema Warna Biru-Ungu** ✅

Gradient color scheme yang kohesif:

- **ColorNeonCyan** (`#00FFFF`) - Bright cyan untuk primary accent
- **ColorNeonBlue** (`#00AFFF`) - Electric blue
- **ColorNeonPurple** (`#AF87FF`) - Bright purple
- **ColorNeonMagenta** (`#FF00FF`) - Magenta untuk highlight
- **ColorDeepPurple** (`#8700FF`) - Deep purple untuk shadow
- **ColorDeepBlue** (`#0087FF`) - Deep blue

#### Gradient pada Title (Q U I B I T):

```
Q → Neon Cyan
U → Neon Blue
I → Neon Purple
B → Neon Magenta
I → Deep Purple
T → Neon Blue
```

### 3. **Hilangkan Box Kedua** ✅

- AppHeader sekarang tampil tanpa border
- Hanya colorful text dengan gradient
- Divider menggunakan heavy line (`━`) dengan warna neon cyan
- Lebih clean dan tidak mengganggu

## 📁 File yang Dimodifikasi

### 1. `internal/tui/border.go` (NEW)

**Sistem Border Professional**

- Unicode box-drawing characters
- Multiple border styles: Single, Double, Heavy, Rounded
- Reusable Box component untuk framed content
- Support untuk title pada border
- Responsive terhadap terminal width

### 2. `internal/tui/design_system.go`

**Enhanced Color Palette**

```go
// Blue-Purple Theme Colors
ColorNeonCyan    = "1;38;5;51"   // Primary accent
ColorNeonBlue    = "1;38;5;39"   // Electric blue
ColorNeonPurple  = "1;38;5;141"  // Bright purple
ColorNeonMagenta = "1;38;5;201"  // Highlight
ColorDeepBlue    = "1;38;5;33"   // Deep blue
ColorDeepPurple  = "1;38;5;93"   // Deep purple (shadow)
```

### 3. `internal/tui/splash.go`

**3D Splash Screen**

- Function `splashHeroTitleLinesColorful()` - Multi-color gradient title
- Function `extrudeBlockStyledBluePurple()` - 3D extrusion dengan shadow
- Function `applyGradientToLine()` - Gradient color application
- Clear screen pada startup untuk dramatic entrance
- Bordered splash box dengan double-line style

### 4. `internal/tui/ui.go`

**Simplified UI**

- AppHeader tanpa box border
- Enhanced Heading dengan visual indicator (`▸`)
- Enhanced Status dengan diamond indicator (`◆`)
- Enhanced Done dengan checkmark (`✓`)
- Divider dengan heavy line dan neon cyan

### 5. `internal/tui/selector.go`

**Enhanced Selector**

- Visual indicator `▸` untuk selected item (neon cyan)
- Group headers dengan neon purple
- Better hint formatting

### 6. `internal/tui/motion.go`

**Dynamic Spinner**

- Braille pattern spinner: `⠋ ⠙ ⠹ ⠸ ⠼ ⠴ ⠦ ⠧ ⠇ ⠏`
- Faster animation (80ms interval)
- Neon blue color untuk spinner frame
- Status text dengan softer blue

## 🎯 Quality Bar Achieved

### ✅ Professional Appearance

- Border system yang rapi dan konsisten
- 3D effect seperti HEXSEC GPT reference
- Cohesive blue-purple color theme

### ✅ High-End Feel

- Vibrant neon colors yang tidak generic
- Smooth animations dan transitions
- Responsive terhadap terminal size

### ✅ Hacker-Grade Aesthetic

- Clean, modern, technical look
- Professional typography dengan visual indicators
- Advanced terminal features (box-drawing, Braille, ANSI colors)

## 🚀 Usage

### Build

```bash
go build -o quibit .
```

### Run

```bash
./quibit generate
```

### Test dengan terminal width berbeda

```bash
COLUMNS=120 ./quibit generate  # Full 3D effect
COLUMNS=80 ./quibit generate   # Standard
```

### Disable splash (jika diperlukan)

```bash
./quibit --no-splash generate
```

### Disable animations

```bash
./quibit --no-anim generate
```

## 🎨 Visual Examples

### Splash Screen Hierarchy

```
╔═══════════════════════════════════════╗
║                                       ║
║    3D TITLE dengan GRADIENT COLORS    ║
║         (Cyan→Blue→Purple)            ║
║                                       ║
║          Tagline Context              ║
║                                       ║
║        by Albert Mangiri              ║
║                                       ║
║           version info                ║
║                                       ║
╚═══════════════════════════════════════╝
```

### App Header (tanpa box)

```
  QUIBIT  (colorful gradient)
  Intelligent project generator for engineers.
  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

### Selection Menu

```
  Select a mode.
  ▸ Generate project     ← Selected (Neon Cyan)
    Continue project     ← Normal (Light Gray)
    View saved projects
    Quit

  ↑/↓ navigate · Enter select
```

## 🔧 Customization

### Ubah Warna Theme

Edit `internal/tui/design_system.go`:

```go
const (
    ColorNeonCyan = "1;38;5;51"  // Ganti dengan color code lain
    // ...
)
```

### Ubah 3D Shadow Depth

Edit di `internal/tui/splash.go`, function `splashHeroTitleLinesColorful()`:

```go
big3D := extrudeBlockStyledBluePurple(big, 2, 1)
//                                         ↑  ↑
//                                        dx  dy
//                                    (horizontal, vertical)
```

### Ubah Border Style

Pilih border style di `internal/tui/border.go`:

- `BorderSingle` - Simple single line
- `BorderDouble` - ╔═══╗ (current untuk splash)
- `BorderHeavy` - ┏━━━┓
- `BorderRounded` - ╭───╮

## 📊 Technical Details

### Color Codes (ANSI 256)

- `51` = Bright Cyan (#00FFFF)
- `39` = Electric Blue (#00AFFF)
- `141` = Bright Purple (#AF87FF)
- `201` = Magenta (#FF00FF)
- `93` = Deep Purple (#8700FF)

### Unicode Characters

- Box Drawing: `┌─┐│└─┘` (Single), `╔═╗║╚═╝` (Double), `┏━┓┃┗━┛` (Heavy)
- Indicators: `▸` (triangle right), `◆` (diamond), `✓` (checkmark)
- Braille: `⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏` (spinner patterns)

## 🎯 Design Philosophy

1. **Konsistensi**: Semua elemen UI menggunakan design tokens yang sama
2. **Hierarchy**: Visual weight yang jelas (title > heading > body > muted)
3. **Feedback**: Clear visual feedback untuk setiap action
4. **Responsiveness**: Adaptive terhadap terminal size
5. **Accessibility**: Graceful degradation untuk NO_COLOR environment

## 📝 Notes

- Warna akan muncul dengan benar di terminal yang support 256 colors
- Set `NO_COLOR=1` environment variable untuk disable colors
- 3D effect terbaik pada terminal width ≥ 80 columns
- Tested pada Linux terminal (xterm, gnome-terminal, etc)

---

_Last updated: 2026-02-06_
_Version: 1.0.0_
