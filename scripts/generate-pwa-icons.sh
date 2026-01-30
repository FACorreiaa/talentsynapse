#!/bin/bash
# generate-pwa-icons.sh - Generate PNG icons from SVG for PWA

set -e

ICON_SVG="assets/static/icon.svg"
OUTPUT_DIR="assets/static"

echo "🎨 Generating PWA icons..."

# Check if ImageMagick is installed
if command -v convert &> /dev/null; then
    echo "✅ Using ImageMagick"
    convert "$ICON_SVG" -resize 192x192 "$OUTPUT_DIR/icon-192.png"
    convert "$ICON_SVG" -resize 512x512 "$OUTPUT_DIR/icon-512.png"
    echo "✅ Generated: icon-192.png (192x192)"
    echo "✅ Generated: icon-512.png (512x512)"
# Check if Inkscape is installed
elif command -v inkscape &> /dev/null; then
    echo "✅ Using Inkscape"
    inkscape "$ICON_SVG" -w 192 -h 192 -o "$OUTPUT_DIR/icon-192.png"
    inkscape "$ICON_SVG" -w 512 -h 512 -o "$OUTPUT_DIR/icon-512.png"
    echo "✅ Generated: icon-192.png (192x192)"
    echo "✅ Generated: icon-512.png (512x512)"
# Check if rsvg-convert is installed (from librsvg)
elif command -v rsvg-convert &> /dev/null; then
    echo "✅ Using rsvg-convert"
    rsvg-convert -w 192 -h 192 "$ICON_SVG" -o "$OUTPUT_DIR/icon-192.png"
    rsvg-convert -w 512 -h 512 "$ICON_SVG" -o "$OUTPUT_DIR/icon-512.png"
    echo "✅ Generated: icon-192.png (192x192)"
    echo "✅ Generated: icon-512.png (512x512)"
else
    echo "❌ No icon converter found!"
    echo ""
    echo "Please install one of the following:"
    echo "  • ImageMagick:  brew install imagemagick  (or apt install imagemagick)"
    echo "  • Inkscape:     brew install --cask inkscape"
    echo "  • librsvg:      brew install librsvg  (or apt install librsvg2-bin)"
    echo ""
    echo "Or use an online converter:"
    echo "  • https://cloudconvert.com/svg-to-png"
    echo "  • https://svgtopng.com/"
    exit 1
fi

echo ""
echo "🎉 Done! Generated PWA icons:"
ls -lh "$OUTPUT_DIR"/icon-*.png

echo ""
echo "📱 Next steps:"
echo "1. Commit and push: git add assets/static/*.png && git commit -m 'Add PWA PNG icons'"
echo "2. Deploy to server"
echo "3. On mobile: Clear cache and reinstall PWA"
