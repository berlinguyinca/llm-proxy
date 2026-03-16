#!/bin/bash
set -e

echo "🚀 Installing LLM Proxy..."

# Check if LM Studio is installed, install if not
if ! command -v lmstudio &> /dev/null; then
    echo "Installing LM Studio from Brew..."
    brew install lmstudio 2>/dev/null || {
        echo "⚠️  Could not auto-install LM Studio. Visit: https://lmstudio.ai"
        read -p "Press Enter to continue without LM Studio (will skip model auto-discovery)..."
    }
fi

# Check for go installation
if ! command -v go &> /dev/null; then
    echo "Installing Go 1.22+"
    if brew versions go | grep -q "go.*1\.[23]"; then
        brew upgrade go
    else
        brew install go@1.22
    fi
fi

# Check for docker (optional but recommended)
if ! command -v docker &> /dev/null; then
    echo "⚠️  Docker not found. Some features require Docker for full deployment."
else
    read -p "Install Docker Desktop? (Y/n) " -n 1 -r; echo ""
    case $REPLY in [Nn]*|[*]) 
        echo "Docker installation skipped." 
        ;;
    *)
        brew install --cask docker-desktop &>/dev/null || true
        ;;
    esac
fi

echo ""
read -p "1) Download from GitHub (recommended)? (y/N) " -n 1 -r; echo ""
if [[ $REPLY =~ ^[Yy]$ ]]; then
    curl -fsSL https://github.com/berlinguyinca/llm-proxy/releases/latest/download/installer.sh -o /tmp/installer.sh
    chmod +x /tmp/installer.sh
    bash /tmp/installer.sh
    rm /tmp/installer.sh
else
    # Manual installation
    echo "Building LLM Proxy server..."
    go install github.com/berlinguyinca/llm-proxy/cmd/proxy@latest
    
    echo "Building CLI manager..."
    go install github.com/berlinguyinca/llm-proxy/cmd/management@latest
    
    echo ""
    echo "✓ LLM Proxy installed successfully!"
    echo ""
    echo "To configure:"
    echo "  cp $HOME/.config/opencode/models.yaml.example $HOME/.config/opencode/models.yaml"
    echo ""
    echo "Or use Docker (recommended):"
    echo "  docker-compose up -d --build"
fi

# Ask about running as service
read -p "Run LLM Proxy as a background service? (Y/n) " -n 1 -r; echo ""
if [[ $REPLY =~ ^[Yy]$ ]]; then
    if command -v lmstudio &> /dev/null && command -v proxy &> /dev/null; then
        ln -sf "$(which proxy)" "$HOME/Library/LaunchAgents/com.llm-proxy.server"
        cat > "$HOME/Library/LaunchAgents/com.llm-proxy.server.plist" << 'PLIST'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
    <dict>
        <key>Label</key>
        <string>com.llm-proxy.server</string>
        <key>ProgramArguments</key>
        <array>
            <string>/opt/homebrew/bin/proxy</string>
            <string>--config</string>
            <string>$HOME/.config/llm-proxy/config.yaml</string>
            <string>--port</string>
            <string>9999</string>
        </array>
        <key>RunAtLoad</key>
        <true/>
        <key>KeepAlive</key>
        <false/>
        <key>StandardOutPath</key>
        <string>/tmp/llm-proxy.log</string>
        <key>StandardErrorPath</key>
        <string>/tmp/llm-proxy-error.log</string>
    </dict>
</plist>
PLIST
        echo "✓ Service created. Start with: launchctl load $HOME/Library/LaunchAgents/com.llm-proxy.server.plist"
    else
        echo "⚠️  Cannot create service - requires proper binary installation via installer."
    fi
fi

echo ""
echo "========================================="
echo "   LLM Proxy Installation Complete!     "
echo "========================================="
echo ""
echo "Next steps:"
echo "1. Start LM Studio: lmstudio"
echo "2. Install Go CLI tools (if not done): go install github.com/berlinguyinca/llm-proxy/cmd/@latest"
echo "3. Configure at: $HOME/.config/llm-proxy/config.yaml"
echo "4. Or use Docker: docker-compose up -d --build"
echo "5. Check health: curl http://localhost:9999/health"
echo "6. List models: llm-proxy-manager models list"
echo ""
echo "Visit: https://github.com/berlinguyinca/llm-proxy"
