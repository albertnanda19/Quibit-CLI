#!/bin/bash

# Demo script to show the colorful 3D splash screen
# with proper terminal width

echo "=== QUIBIT CLI - ENHANCED UI DEMO ==="
echo ""
echo "Testing with different terminal widths:"
echo ""

echo "1. Standard width (80 columns):"
COLUMNS=80 ./quibit --help 2>&1 | head -20
echo ""

echo "2. Wide width (100 columns) - Better 3D effect:"
COLUMNS=100 ./quibit --help 2>&1 | head -20
echo ""

echo "3. Extra wide (120 columns) - Full 3D effect:"
COLUMNS=120 ./quibit --help 2>&1 | head -20
echo ""

echo "=== Demo Complete ==="
echo ""
echo "Features implemented:"
echo "  ✓ 3D extruded ASCII art title"
echo "  ✓ Blue-Purple color theme (Cyan, Blue, Purple, Magenta)"
echo "  ✓ Professional bordered splash screen"
echo "  ✓ Simplified app header (removed second box)"
echo "  ✓ Enhanced UI elements (headings, status, selectors)"
echo "  ✓ Dynamic spinner with Braille characters"
echo ""
