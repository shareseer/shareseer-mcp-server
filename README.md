# ShareSeer MCP Server

A Model Context Protocol (MCP) server that provides access to ShareSeer's comprehensive SEC filings, insider transactions, and financial data through Claude and other MCP-compatible AI assistants.

## 🚀 Quick Start

### Installation

#### Option 1: One-line install (Recommended)
```bash
curl -sSL https://raw.githubusercontent.com/shareseer/mcp-server/main/install.sh | sh
```

#### Option 2: Download binary directly
Visit our [releases page](https://github.com/shareseer/mcp-server/releases/latest) and download the appropriate binary for your platform.

### Get Your API Key

1. Sign up at [shareseer.com](https://shareseer.com)
2. Go to your [profile page](https://shareseer.com/profile)
3. Your API key will be displayed (starts with `sk-shareseer-`)

### Claude Desktop Integration

Add to your Claude Desktop configuration (`claude_desktop_config.json`):

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

**Configuration file locations:**
- **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`

## 📊 Available Tools

### Company Information
- **`get_company_info`** - Get basic company information by ticker
- **`get_company_filings`** - Get recent SEC filings for a specific company
- **`get_insider_transactions`** - Get insider trading transactions for a company

### Market-Wide Data
- **`get_recent_filings`** - Get recent SEC filings across all companies
- **`get_recent_insider_activity`** - Get recent insider trading activity

### Largest Transactions (Premium Feature)
- **`get_largest_daily_transactions`** - Get largest daily insider transactions
- **`get_largest_weekly_transactions`** - Get largest weekly insider transactions

## 💎 Subscription Tiers

| Feature | Free | Premium ($29/mo) | Pro ($99/mo) |
|---------|------|------------------|--------------|
| **Rate Limits** | 10/hour, 50/day | 100/hour, 1K/day | 1K/hour, 10K/day |
| **Data History** | 6 months | 10 years | Unlimited |
| **Company Data** | ✅ Basic info & filings | ✅ All features | ✅ All features |
| **Insider Transactions** | ✅ Limited (3 results) | ✅ Unlimited | ✅ Unlimited |
| **Largest Transactions** | ✅ Current week only | ✅ Historical data | ✅ Full access |
| **Pagination** | ❌ | ✅ | ✅ |
| **Support** | Community | Email | Priority |

[**Upgrade to Premium**](https://shareseer.com/upgrade?source=mcp) | [**Go Pro**](https://shareseer.com/upgrade?plan=pro&source=mcp)

## 🔧 Usage Examples

### Get Company Information
Ask Claude: *"What's the latest information about Apple?"*

### Get Recent Insider Transactions  
Ask Claude: *"Show me recent insider trading for Tesla"*

### Get Largest Daily Buyers
Ask Claude: *"Who made the biggest stock purchases today?"*

### Get Largest Weekly Sellers
Ask Claude: *"What were the largest insider sales last week?"*

## 🛠️ For Developers

### Building from Source

```bash
git clone https://github.com/shareseer/mcp-server.git
cd mcp-server
make build
```

### Running the Server

```bash
# Start the server
./shareseer-mcp

# Server will be available at http://localhost:8081
# API info at http://localhost:8081/mcp/info
```

### Configuration

The server can be configured via environment variables:

```bash
export SHARESEER_API_KEY="your-api-key"
export SHARESEER_HOST="0.0.0.0"
export SHARESEER_PORT="8081"
```

## 🔐 Security & Privacy

- **No sensitive data exposure** - Internal identifiers are not exposed
- **Rate limiting** - Prevents abuse with tier-based limits
- **Read-only access** - Server only reads data, never modifies
- **Secure authentication** - API keys are validated against ShareSeer's user database

## 🐛 Troubleshooting

### Common Issues

**"Invalid API key" error:**
- Verify your API key starts with `sk-shareseer-`
- Check that your ShareSeer account is active
- Ensure the API key is set in your environment

**"Rate limit exceeded" error:**
- Check your current tier limits
- Wait for the rate limit window to reset
- Consider upgrading to Premium for higher limits

### Getting Help

- **Free users**: [GitHub Issues](https://github.com/shareseer/mcp-server/issues)
- **Premium/Pro users**: Email support@shareseer.com
- **Documentation**: [ShareSeer MCP Docs](https://shareseer.com/docs/mcp)

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Support

Love the ShareSeer MCP server? Here's how you can help:

- ⭐ **Star this repo** on GitHub
- 🐛 **Report bugs** via GitHub Issues  
- 💡 **Request features** via GitHub Discussions
- 📢 **Share** with other developers and traders
- 💎 **Upgrade to Premium** to support continued development

---

**Built with ❤️ by the ShareSeer team**

[Website](https://shareseer.com) • [Twitter](https://twitter.com/shareseer) • [Email](mailto:support@shareseer.com)

## ⚠️ Note for Developers

This repository contains the public MCP server interface. The actual data access implementation connects to ShareSeer's proprietary data infrastructure. To run this server, you need:

1. A valid ShareSeer API key
2. Access to ShareSeer's data endpoints
3. Proper authentication credentials

The server acts as a bridge between the MCP protocol and ShareSeer's financial data services.