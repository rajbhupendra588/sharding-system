# OAuth Social Login Configuration Guide

## Quick Setup

OAuth providers are configured via **environment variables**. You need to:

1. **Get OAuth credentials** from the provider (Google, GitHub, or Facebook)
2. **Set environment variables** before starting the manager server
3. **Restart the manager server**

## Step-by-Step Instructions

### Option 1: Set Environment Variables in Terminal

```bash
# For Google OAuth
export GOOGLE_OAUTH_CLIENT_ID="your_google_client_id_here"
export GOOGLE_OAUTH_CLIENT_SECRET="your_google_client_secret_here"

# For GitHub OAuth
export GITHUB_OAUTH_CLIENT_ID="your_github_client_id_here"
export GITHUB_OAUTH_CLIENT_SECRET="your_github_client_secret_here"

# For Facebook OAuth
export FACEBOOK_OAUTH_CLIENT_ID="your_facebook_client_id_here"
export FACEBOOK_OAUTH_CLIENT_SECRET="your_facebook_client_secret_here"

# Optional: Set base URL for OAuth callbacks (auto-detected if not set)
export BASE_URL="http://localhost:8081"

# Then start the manager
./bin/manager
# OR
go run ./cmd/manager
```

### Option 2: Create a `.env` file (if using a process manager)

Create a file called `.env` in the project root:

```bash
GOOGLE_OAUTH_CLIENT_ID=your_google_client_id_here
GOOGLE_OAUTH_CLIENT_SECRET=your_google_client_secret_here
GITHUB_OAUTH_CLIENT_ID=your_github_client_id_here
GITHUB_OAUTH_CLIENT_SECRET=your_github_client_secret_here
FACEBOOK_OAUTH_CLIENT_ID=your_facebook_client_id_here
FACEBOOK_OAUTH_CLIENT_SECRET=your_facebook_client_secret_here
BASE_URL=http://localhost:8081
```

Then source it before starting:
```bash
source .env
./bin/manager
```

### Option 3: Update start.sh script

The `scripts/start.sh` script can be modified to include OAuth variables.

## How to Get OAuth Credentials

### Google OAuth Setup

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Enable **Google+ API** or **Google Identity API**
4. Go to **Credentials** → **Create Credentials** → **OAuth 2.0 Client ID**
5. Application type: **Web application**
6. **Authorized redirect URIs**: 
   - `http://localhost:8081/api/v1/auth/oauth/google/callback` (for local dev)
   - `https://yourdomain.com/api/v1/auth/oauth/google/callback` (for production)
7. Copy the **Client ID** and **Client Secret**

### GitHub OAuth Setup (Detailed Step-by-Step)

#### Step 1: Access GitHub Developer Settings

1. Log in to your GitHub account
2. Click on your **profile picture** (top right corner)
3. Select **Settings** from the dropdown menu
4. Scroll down in the left sidebar and click **Developer settings**
5. Click **OAuth Apps** in the left sidebar (under "Personal access tokens")

#### Step 2: Create a New OAuth App

1. Click the **New OAuth App** button (top right, green button)
2. You'll see a form to fill in your application details

#### Step 3: Fill in Application Details

Fill in the following fields:

**Application name:**
- Enter a descriptive name for your application
- Example: `Sharding System` or `My Sharding App`
- This name will be shown to users when they authorize your app

**Homepage URL:**
- For local development: `http://localhost:3000`
- For production: `https://yourdomain.com` (your actual domain)
- This is where your application is hosted

**Application description (optional):**
- Add a brief description of what your application does
- Example: `Database sharding management system`

**Authorization callback URL:**
- **This is the most important field!**
- For local development: `http://localhost:8081/api/v1/auth/oauth/github/callback`
- For production: `https://yourdomain.com/api/v1/auth/oauth/github/callback`
- **Important:** The URL must match exactly, including the protocol (http/https) and port
- Make sure there's no trailing slash

#### Step 4: Register the Application

1. Click the **Register application** button (green button at the bottom)
2. GitHub will create your OAuth app and redirect you to the app's settings page

#### Step 5: Get Your Credentials

After registration, you'll see your OAuth app's details page:

1. **Client ID:**
   - This is displayed immediately on the page
   - It's a long string like: `Iv1.8a61f9b3a7aba766`
   - Copy this value - you'll need it for `GITHUB_OAUTH_CLIENT_ID`

2. **Client Secret:**
   - Click the **Generate a new client secret** button
   - GitHub will show you the secret (it looks like: `a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6`)
   - **Important:** Copy this immediately! You can only see it once
   - If you lose it, you'll need to generate a new one
   - This is your `GITHUB_OAUTH_CLIENT_SECRET`

#### Step 6: Configure Your Application

1. **Set Environment Variables:**
   ```bash
   export GITHUB_OAUTH_CLIENT_ID="Iv1.8a61f9b3a7aba766"
   export GITHUB_OAUTH_CLIENT_SECRET="a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6"
   ```

2. **Or add to `.env` file:**
   ```bash
   GITHUB_OAUTH_CLIENT_ID=Iv1.8a61f9b3a7aba766
   GITHUB_OAUTH_CLIENT_SECRET=a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6
   ```

#### Step 7: Restart Your Manager Server

1. Stop the current manager server (if running)
2. Start it again:
   ```bash
   ./bin/manager
   # OR
   ./scripts/start.sh
   ```

3. Check the logs to verify GitHub OAuth is configured:
   ```bash
   tail -f logs/manager.log | grep -i oauth
   ```
   
   You should see:
   ```
   {"level":"info","msg":"GitHub OAuth configured","client_id":"Iv1...."}
   {"level":"info","msg":"OAuth social login enabled","providers":1}
   ```

#### Step 8: Test GitHub Login

1. Open your application's login page in a browser
2. You should see a GitHub button (🐙 icon)
3. Click the GitHub button
4. You'll be redirected to GitHub's authorization page
5. Click **Authorize** (or **Authorize [your-app-name]**)
6. You'll be redirected back to your application and logged in

#### Troubleshooting GitHub OAuth

**Problem: "redirect_uri_mismatch" error**
- **Solution:** Make sure the callback URL in GitHub matches exactly:
  - Check for trailing slashes
  - Verify protocol (http vs https)
  - Verify port number (8081 for manager)
  - The URL should be: `http://localhost:8081/api/v1/auth/oauth/github/callback`

**Problem: "Invalid client" error**
- **Solution:** 
  - Verify your Client ID and Client Secret are correct
  - Make sure you copied the entire Client Secret (they're long!)
  - Check that environment variables are set correctly: `echo $GITHUB_OAUTH_CLIENT_ID`

**Problem: GitHub button not showing**
- **Solution:**
  - Check manager logs for OAuth configuration messages
  - Verify environment variables are set before starting manager
  - Restart the manager server after setting variables
  - Check browser console for API errors

**Problem: "Application suspended"**
- **Solution:** 
  - Check your GitHub OAuth app settings
  - Make sure the app is not suspended
  - Verify your GitHub account is in good standing

#### Additional GitHub OAuth Settings

**Update OAuth App Settings:**
- Go back to GitHub → Settings → Developer settings → OAuth Apps
- Click on your app name
- You can update the callback URL, homepage URL, or app name anytime
- **Note:** Changing the callback URL requires updating your application configuration

**Multiple Environments:**
If you need different OAuth apps for development and production:
1. Create separate OAuth apps for each environment
2. Use different environment variables for each
3. Development: `http://localhost:8081/api/v1/auth/oauth/github/callback`
4. Production: `https://yourdomain.com/api/v1/auth/oauth/github/callback`

**Permissions and Scopes:**
- The default scope `user:email` is automatically requested
- This allows reading the user's email address
- No additional scopes are required for basic login functionality

### Facebook OAuth Setup

1. Go to [Facebook Developers](https://developers.facebook.com/)
2. Create a new app or select an existing one
3. Add **Facebook Login** product
4. Go to **Settings** → **Basic**
5. Add **Valid OAuth Redirect URIs**:
   - `http://localhost:8081/api/v1/auth/oauth/facebook/callback` (for local dev)
   - `https://yourdomain.com/api/v1/auth/oauth/facebook/callback` (for production)
6. Go to **Settings** → **Basic** to find **App ID** and **App Secret**

## Testing

After setting the environment variables and restarting the manager:

1. Open the login page in your browser
2. You should see social login buttons for the configured providers
3. Click a button to test the OAuth flow

## Troubleshooting

- **Buttons not showing**: Make sure environment variables are set and manager is restarted
- **OAuth error**: Check that redirect URIs match exactly in provider settings
- **"Invalid client"**: Verify Client ID and Secret are correct
- **Check logs**: Look at `logs/manager.log` for OAuth-related errors

