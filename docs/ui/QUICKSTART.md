# UI Quick Start Guide

This guide will help you get the Sharding System UI up and running quickly.

## Prerequisites

1. **Node.js 18+** installed on your system
2. **Running Sharding System Backend**:
   - Manager service on port 8081
   - Router service on port 8080

## Installation Steps

### 1. Install Dependencies

```bash
cd ui
npm install
```

### 2. Start Development Server

```bash
npm run dev
```

The UI will be available at `http://localhost:3000`

### 3. Access the UI

Open your browser and navigate to `http://localhost:3000`

## First Steps

### 1. Check System Health

- Navigate to the **Health** page from the sidebar
- Verify that both Manager and Router services show "healthy" status

### 2. View Existing Shards

- Go to the **Shards** page
- You should see a list of all configured shards
- If no shards exist, you'll see an empty state

### 3. Create Your First Shard

1. Click the **"Create Shard"** button on the Shards page
2. Fill in the form:
   - **Name**: A descriptive name (e.g., `shard-01`)
   - **Primary Endpoint**: PostgreSQL connection string (e.g., `postgres://user:pass@host:5432/db`)
   - **Virtual Node Count**: Number of virtual nodes (default: 256)
   - **Replicas**: Add replica endpoints if available
3. Click **"Create Shard"**

### 4. Execute a Query

1. Navigate to **Query Executor** from the sidebar
2. Enter:
   - **Shard Key**: A key to route the query (e.g., `user-123`)
   - **SQL Query**: Your SQL query (e.g., `SELECT * FROM users WHERE id = $1`)
   - **Parameters**: Comma-separated values (e.g., `user-123`)
   - **Consistency Level**: Choose `strong` or `eventual`
3. Click **"Execute Query"**
4. View results in the results panel

### 5. Monitor Health

- Go to the **Health** page to see:
  - System-wide health status
  - Individual service health
  - Shard health summary

### 6. View Metrics

- Navigate to **Metrics** to see:
  - Query rate over time
  - Average latency
  - Shard performance

## Configuration

### API Endpoints

By default, the UI connects to:
- Manager API: `http://localhost:8081`
- Router API: `http://localhost:8080`

To change these:
1. Go to **Settings** page
2. Update the API URLs
3. Click **"Save Settings"**

### Environment Variables

You can also configure endpoints via environment variables:

```bash
# Create .env file
VITE_MANAGER_URL=http://localhost:8081
VITE_ROUTER_URL=http://localhost:8080
```

## Common Tasks

### Split a Shard

1. Go to **Resharding** page
2. Click **"Split Shard"**
3. Select source shard
4. Configure target shards
5. Click **"Start Split"**
6. Monitor progress on the job detail page

### Merge Shards

1. Go to **Resharding** page
2. Click **"Merge Shards"**
3. Select 2+ source shards
4. Configure target shard
5. Click **"Start Merge"**
6. Monitor progress on the job detail page

### Promote a Replica

1. Go to **Shards** page
2. Click on a shard to view details
3. Find the replica you want to promote
4. Click **"Promote"** button
5. Confirm the action

## Troubleshooting

### UI Won't Start

- Check Node.js version: `node --version` (should be 18+)
- Clear node_modules and reinstall: `rm -rf node_modules && npm install`
- Check port 3000 is available

### Can't Connect to Backend

- Verify backend services are running
- Check API URLs in Settings
- Verify CORS is configured on backend
- Check browser network tab for errors

### No Data Showing

- Verify backend services are healthy (check Health page)
- Check browser network tab for API errors
- Verify shards exist in the system

### Build Errors

- Run `npm run type-check` to check TypeScript errors
- Run `npm run lint` to check linting errors
- Clear build cache: `rm -rf dist node_modules/.vite`

## Next Steps

- Explore all features in the UI
- Read the [Architecture Guide](./ARCHITECTURE.md) for detailed documentation
- Check the main project documentation for backend setup

## Support

For issues or questions:
1. Check the browser console for errors
2. Review backend logs
3. Consult the main project documentation

