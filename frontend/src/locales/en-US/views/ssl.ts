// SSL page texts
export default {
  title: 'SSL Certificates',
  subtitle: 'Apply, renew and manage HTTPS certificates',
  provider: {
    manual: 'Manual Upload'
  },
  status: {
    valid: 'Valid',
    expiring: 'Expiring Soon',
    expired: 'Expired',
    pending: 'Issuing'
  },
  columns: {
    domain: 'Domain',
    provider: 'Provider',
    expireAt: 'Expires At',
    status: 'Status',
    actions: 'Actions'
  },
  expireText: {
    expired: 'Expired {days} days ago',
    remaining: '{days} days remaining'
  },
  actions: {
    apply: 'Apply Certificate',
    detail: 'Details',
    renew: 'Renew',
    delete: 'Delete'
  },
  detail: {
    title: 'Certificate Details',
    providerLabel: 'Provider',
    issuer: 'Issuer',
    email: 'Contact Email'
  },
  messages: {
    loadListFailed: 'Failed to load the SSL certificate list',
    renewSuccess: 'Certificate renewed',
    renewPending: 'Auto-renewal is not implemented yet: the server only records the request and does not actually issue the certificate. Upload a certificate manually for now.',
    emailRequired: 'Enter an email in the application form before renewing',
    renewFailed: 'Renew failed',
    applySuccess: 'Certificate application submitted',
    applyFailed: 'Apply failed',
    deleteSuccess: 'Certificate deleted',
    deleteFailed: 'Delete failed'
  },
  rules: {
    selectSite: 'Please select a site',
    email: 'Please enter an email address',
    cert: 'Please enter the certificate content',
    key: 'Please enter the private key content',
    dnsProvider: 'Please select a DNS provider'
  },
  apply: {
    title: 'Apply SSL Certificate',
    selectSite: 'Select Site',
    sitePlaceholder: 'Select the site to bind the certificate to',
    methodLabel: 'Application Method',
    letsencryptOption: "Let's Encrypt (auto issue)",
    emailLabel: 'Email Address',
    emailPlaceholder: 'Used to receive certificate expiry reminders',
    verificationLabel: 'Verification Method',
    httpOption: 'HTTP Verification',
    dnsOption: 'DNS Verification',
    verificationNoteLabel: 'Verification Note',
    httpNote: 'A verification file will be placed in the site directory automatically to complete HTTP-01 verification',
    dnsProviderLabel: 'DNS Provider',
    dnsProviderPlaceholder: 'Select a configured DNS provider account',
    dnsNote: 'A TXT record will be added automatically to complete DNS-01 verification; no manual action needed',
    certLabel: 'Certificate (PEM)',
    keyLabel: 'Private Key (PEM)',
    dnsAccountLabel: '{name} ({type})'
  },
  applyStatus: {
    dnsAdding: 'Adding DNS TXT record...'
  },
  confirm: {
    deleteTitle: 'Delete Certificate',
    deleteContent: 'Delete the SSL certificate for domain "{domain}"? The site will no longer be able to use HTTPS after deletion.',
    deleteConfirmText: 'Delete'
  }
}
