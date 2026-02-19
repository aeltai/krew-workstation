# Create Krew Workstation GitHub Repo

The GitHub API token doesn't have repo creation permission. Create the repo manually:

## 1. Create the repo on GitHub

1. Go to https://github.com/new
2. Repository name: `krew-workstation`
3. Description: `Krew Workstation - kubectl plugin manager for Rancher Dashboard`
4. Public
5. Add a README (optional)
6. Create repository

## 2. Push the code

From the `ui-plugin-examples` repo root:

```bash
# Create a new orphan branch with only krew content
git subtree split -P pkg/krew -b krew-workstation

# Add the new remote
git remote add krew-workstation git@github.com:aeltai/krew-workstation.git

# Push (replace aeltai with your username if different)
git push krew-workstation krew-workstation:main
```

Or copy the folder and push:

```bash
# Clone your new empty repo
git clone git@github.com:aeltai/krew-workstation.git
cd krew-workstation

# Copy krew content (from ui-plugin-examples root)
cp -r ../ui-plugin-examples/pkg/krew/* .
# Or: rsync -av --exclude=node_modules --exclude=dist ../ui-plugin-examples/pkg/krew/ .

# Commit and push
git add .
git commit -m "Initial Krew Workstation"
git push origin main
```

## 3. Repo structure

The repo should contain:
- `backend/` - Go backend
- `helm/krew-workstation/` - Helm chart
- `helm/README.md` - Helm install instructions
- `KrewPage.vue`, `index.ts`, `product.ts` - Rancher UI extension
- `docker-compose.yml` - Local dev
- `README.md` - Project docs
