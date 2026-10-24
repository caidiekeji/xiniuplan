// deploy page texts
export default {
  pageTitle: 'Git Deploy',
  pageSubtitle: 'Pull, build and publish triggered by repo Webhook, or deploy manually from the panel',
  refresh: 'Refresh',
  addApp: 'Add App',
  buildTypes: {
    static: 'Static Site',
    node: 'Node Build',
    dockerfile: 'Dockerfile'
  },
  status: {
    success: 'Success',
    failed: 'Failed',
    deploying: 'Deploying'
  },
  columns: {
    name: 'App Name',
    repo: 'Repository',
    branch: 'Branch',
    buildType: 'Build Type',
    destPath: 'Target Directory',
    lastDeploy: 'Last Deploy',
    actions: 'Actions'
  },
  notDeployed: 'Not deployed',
  actions: {
    deploy: 'Deploy',
    logs: 'Logs',
    edit: 'Edit',
    delete: 'Delete'
  },
  messages: {
    deployDone: 'App "{name}" deployed',
    deployFailed: 'Deployment failed',
    updated: 'App updated',
    created: 'App created',
    saveFailed: 'Save failed',
    deleted: 'App deleted',
    deleteFailed: 'Deletion failed'
  },
  rules: {
    nameRequired: 'Please enter the app name',
    repoRequired: 'Please enter the repository URL',
    destPathRequired: 'Please enter the deploy target directory'
  },
  empty: {
    title: 'No deploy apps yet',
    desc: 'Add a Git repository to deploy manually or configure Webhook for automatic publishing'
  },
  webhook: {
    tipPrefix: 'Webhook URL format: ',
    tipSuffix: ', the signature matches the app Secret'
  },
  form: {
    addTitle: 'Add App',
    editTitle: 'Edit App',
    nameLabel: 'App Name',
    namePlaceholder: 'e.g. blog',
    branchLabel: 'Branch',
    repoLabel: 'Repository',
    repoPlaceholder: "https://github.com/user/repo.git (credentials user:token{'@'} allowed)",
    buildTypeLabel: 'Build Type',
    containerPortLabel: 'Container Port',
    outDirLabel: 'Artifact Directory',
    destPathLabel: 'Deploy Target Directory',
    destPathPlaceholder: '/var/www/blog (static/node builds are published here)',
    hint: "Webhook URL: /api/v1/deploy/webhook/{'{id}'}?sign=Secret, configure it in the repository Webhook settings"
  },
  logs: {
    title: 'Deploy Logs · {name}',
    empty: 'No deploy records yet',
    expand: 'Expand',
    collapse: 'Collapse'
  },
  deleteDialog: {
    title: 'Delete App',
    content: 'Delete deploy app "{name}"? The build output directory will not be removed. This action cannot be undone.',
    confirm: 'Delete'
  }
}
