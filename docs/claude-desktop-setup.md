# Claude Desktop Integration Guide

This guide walks you through setting up the ShareSeer MCP server with Claude Desktop for seamless access to SEC filing and insider trading data.

## Prerequisites

- Claude Desktop application installed
- ShareSeer account with API key
- ShareSeer MCP server installed

## Step 1: Install ShareSeer MCP Server

Choose one of the installation methods:

### Quick Install (Recommended)
```bash
curl -sSL https://raw.githubusercontent.com/shareseer/mcp-server/main/install.sh | sh
```

### Manual Download
Download the appropriate binary for your platform from the [releases page](https://github.com/shareseer/mcp-server/releases/latest).

## Step 2: Get Your API Key

1. **Sign up** at [shareseer.com](https://shareseer.com) if you haven't already
2. **Log in** to your account
3. **Go to your profile** at [shareseer.com/profile](https://shareseer.com/profile)
4. **Copy your API key** (starts with `sk-shareseer-`)

## Step 3: Configure Claude Desktop

### Find Your Configuration File

**macOS:**
```bash
~/Library/Application Support/Claude/claude_desktop_config.json
```

**Windows:**
```bash
%APPDATA%\Claude\claude_desktop_config.json
```

### Create/Edit Configuration

Open the configuration file in your preferred text editor and add the ShareSeer MCP server:

```json
{
  "mcpServers": {
    "shareseer": {
      "command": "/usr/local/bin/shareseer-mcp",
      "env": {
        "SHARESEER_API_KEY": "sk-shareseer-your-actual-api-key-here"
      }
    }
  }
}
```

### Configuration Options

#### Basic Configuration (Recommended)
```json
{
  "mcpServers": {
    "shareseer": {
      "command": "/usr/local/bin/shareseer-mcp",
      "env": {
        "SHARESEER_API_KEY": "sk-shareseer-your-api-key-here"
      }
    }
  }
}
```

#### Custom Installation Path
If you installed to a custom location:
```json
{
  "mcpServers": {
    "shareseer": {
      "command": "/path/to/your/shareseer-mcp",
      "env": {
        "SHARESEER_API_KEY": "sk-shareseer-your-api-key-here"
      }
    }
  }
}
```

#### Docker Configuration
If you prefer to run via Docker:
```json
{
  "mcpServers": {
    "shareseer": {
      "command": "docker",
      "args": [
        "run", "--rm", "--network=host",
        "-e", "SHARESEER_API_KEY=sk-shareseer-your-api-key-here",
        "shareseer/mcp-server"
      ]
    }
  }
}
```

#### Multiple Configurations
You can have multiple MCP servers:
```json
{
  "mcpServers": {
    "shareseer": {
      "command": "/usr/local/bin/shareseer-mcp",
      "env": {
        "SHARESEER_API_KEY": "sk-shareseer-your-api-key-here"
      }
    },
    "other-mcp-server": {
      "command": "/path/to/other-server"
    }
  }
}
```

## Step 4: Restart Claude Desktop

After updating the configuration:

1. **Quit Claude Desktop** completely
2. **Restart the application**
3. **Wait a few seconds** for the MCP server to initialize

## Step 5: Test the Integration

Try these example queries in Claude Desktop:

### Company Information
> "What's the latest insider trading activity for Apple (AAPL)?"

### Recent Filings
> "Show me the most recent SEC filings from Tesla"

### Largest Transactions
> "What were the largest insider purchases this week?"

### Market Analysis
> "Give me a summary of recent insider trading trends"

## Troubleshooting

### Common Issues

#### MCP Server Not Loading
**Symptoms:** Claude doesn't recognize ShareSeer tools

**Solutions:**
1. Check that the binary path is correct in your config
2. Verify the binary is executable: `chmod +x /usr/local/bin/shareseer-mcp`
3. Test the binary directly: `/usr/local/bin/shareseer-mcp --help`
4. Check Claude Desktop logs for error messages

#### Invalid API Key Error
**Symptoms:** "Invalid API key" responses

**Solutions:**
1. Verify your API key starts with `sk-shareseer-`
2. Check your ShareSeer account is active
3. Ensure no extra spaces or characters in the config
4. Try regenerating your API key from the profile page

#### Rate Limiting
**Symptoms:** "Rate limit exceeded" messages

**Solutions:**
1. Check your subscription tier limits
2. Wait for the rate limit window to reset
3. Consider upgrading to Premium for higher limits

#### Connection Issues
**Symptoms:** "Connection refused" or timeout errors

**Solutions:**
1. Ensure the MCP server is running
2. Check firewall settings (port 8081)
3. Verify Redis connectivity (if self-hosting)

### Debug Steps

#### Test MCP Server Directly
```bash
# Start the server manually
/usr/local/bin/shareseer-mcp

# In another terminal, test the API
curl http://localhost:8081/mcp/info
```

#### Check Configuration Syntax
```bash
# Validate JSON syntax
python -m json.tool ~/Library/Application\ Support/Claude/claude_desktop_config.json
```

#### View Logs
Claude Desktop logs can help identify configuration issues:

**macOS:**
```bash
tail -f ~/Library/Logs/Claude/claude-desktop.log
```

**Windows:**
```bash
# Check Event Viewer or Claude Desktop console output
```

## Advanced Configuration

### Environment Variables

You can use environment variables for configuration:

```json
{
  "mcpServers": {
    "shareseer": {
      "command": "/usr/local/bin/shareseer-mcp",
      "env": {
        "SHARESEER_API_KEY": "sk-shareseer-your-api-key-here",
        "SHARESEER_HOST": "0.0.0.0",
        "SHARESEER_PORT": "8081",
        "REDIS_ADDR": "localhost:6379"
      }
    }
  }
}
```

### Custom Configuration File

You can specify a custom config file:

```json
{
  "mcpServers": {
    "shareseer": {
      "command": "/usr/local/bin/shareseer-mcp",
      "args": ["--config", "/path/to/custom/config.yaml"],
      "env": {
        "SHARESEER_API_KEY": "sk-shareseer-your-api-key-here"
      }
    }
  }
}
```

## Security Best Practices

1. **Keep your API key secure** - Don't share or commit it to version control
2. **Use environment variables** for sensitive data when possible
3. **Regularly rotate API keys** from your ShareSeer profile
4. **Monitor usage** to detect any unauthorized access
5. **Use the principle of least privilege** - only grant necessary permissions

## Support

- **Documentation:** [ShareSeer MCP Docs](https://github.com/shareseer/mcp-server)
- **Issues:** [GitHub Issues](https://github.com/shareseer/mcp-server/issues)
- **Premium Support:** support@shareseer.com
- **Community:** [Discord](https://discord.gg/shareseer)

## Next Steps

Once you have the integration working:

1. **Explore all available tools** using Claude's natural language interface
2. **Upgrade to Premium** for higher rate limits and advanced features
3. **Share feedback** to help improve the service
4. **Join the community** to share tips and strategies

Happy trading! 📈🚀