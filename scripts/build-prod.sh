#!/bin/bash
set -e

echo "🧹 Cleaning previous builds..."
rm -rf bin/ tmp/
rm -f assets/css/output.css

echo "📝 Generating Templ templates..."
templ generate

echo "🎨 Building optimized CSS..."
tailwindcss -i ./assets/css/index.css -o ./assets/css/output.css --minify

echo "🔨 Building Go binary..."
CGO_ENABLED=0 go build \
  -ldflags="-s -w -X main.version=$(git describe --tags --always --dirty 2>/dev/null || echo 'dev')" \
  -o ./bin/server \
  ./cmd/server

echo ""
echo "📊 Build Statistics:"
echo "===================="
ls -lh bin/server | awk '{print "Binary size:", $5}'
ls -lh assets/css/output.css | awk '{print "CSS size:", $5}'
du -sh assets/ | awk '{print "Total assets:", $1}'

# Optional: UPX compression (uncomment if upx is installed)
# if command -v upx &> /dev/null; then
#     echo ""
#     echo "🗜️  Compressing binary with UPX..."
#     cp bin/server bin/server.uncompressed
#     upx -7 bin/server
#     echo ""
#     echo "Compressed size:"
#     ls -lh bin/server | awk '{print "  Compressed:", $5}'
#     ls -lh bin/server.uncompressed | awk '{print "  Original:", $5}'
# fi

echo ""
echo "✅ Production build complete!"
echo ""
echo "To deploy:"
echo "  scp bin/server user@server:/app/"
echo "  ssh user@server 'DATABASE_URL=... PORT=8080 /app/server'"
