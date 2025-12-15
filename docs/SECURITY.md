# Security Guide

## Lobby Security

The bot uses secure invitation tokens to protect lobbies from unauthorized access. Here's how it works:

### How It Works

1. **Invitation Tokens**: Each lobby has a unique, cryptographically secure token (32 hex characters)
2. **Token Format**: Tokens are displayed as `XXXX-XXXX-XXXX-XXXX` for readability
3. **Group Integration**: If you create a lobby in a Telegram group/channel, it's automatically linked

### Setting Up Securely

#### Option 1: Private Group/Channel (Recommended)

This is the most secure method:

1. **Create a private Telegram group** with just you and your partner
2. **Add the bot** to the group (`@YourBotName`)
3. **Both users run `/start`** in the group
4. The bot automatically detects the group and links the lobby to it
5. Only members of that group can interact with the lobby

**Benefits:**
- No tokens to share manually
- Group membership controls access
- Works seamlessly in your existing private chat

#### Option 2: Direct Invitation Token

If you prefer not to use a group:

1. **One user runs `/start`** to create a lobby
2. **Bot provides a secure token** (e.g., `A1B2-C3D4-E5F6-G7H8`)
3. **Share token privately** with your partner via:
   - Private Telegram message
   - Encrypted messaging app
   - In person
   - **NOT** in public channels/groups
4. **Partner runs `/start <token>`** to join
5. Use `/invite` to view the token again
6. Use `/regenerate_invite` if token is compromised

### Security Best Practices

1. **Never share tokens publicly** - Anyone with the token can join your lobby
2. **Use private groups** - Most secure option, no tokens needed
3. **Regenerate if compromised** - If you suspect someone has your token, use `/regenerate_invite`
4. **Keep tokens private** - Treat tokens like passwords
5. **Verify partner identity** - Make sure you're sharing with the right person

### Commands

- `/start` - Create lobby or join with token
- `/invite` - View your lobby's invitation token
- `/regenerate_invite` - Generate a new token (invalidates old one)

### Token Format

Tokens are 32 hexadecimal characters, displayed as:
```
A1B2-C3D4-E5F6-G7H8-I9J0-K1L2-M3N4-O5P6
```

You can use them with or without dashes:
- `/start A1B2-C3D4-E5F6-G7H8-I9J0-K1L2-M3N4-O5P6`
- `/start A1B2C3D4E5F6G7H8I9J0K1L2M3N4O5P6`

Both formats work the same way.

