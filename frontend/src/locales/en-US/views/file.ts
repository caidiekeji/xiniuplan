// file page messages
export default {
  treeTitle: 'Directory Tree',
  recycleTitle: 'Recycle Bin',
  buttons: {
    mkdir: 'New Folder',
    mkfile: 'New File',
    upload: 'Upload',
    refresh: 'Refresh'
  },
  search: {
    placeholder: 'Search current directory…',
    searching: 'Searching',
    hint: 'Keyword "{keyword}" · {count} match(es)',
    exit: 'Exit search'
  },
  batch: {
    selectedCount: '{n} selected',
    copy: 'Copy',
    move: 'Move',
    compress: 'Compress',
    clear: 'Clear selection'
  },
  columns: {
    name: 'Name',
    size: 'Size',
    mode: 'Permissions',
    modTime: 'Modified',
    actions: 'Actions'
  },
  actions: {
    rename: 'Rename',
    download: 'Download',
    delete: 'Delete',
    more: 'More',
    restore: 'Restore',
    purge: 'Delete permanently',
    remove: 'Remove',
    close: 'Close',
    save: 'Save (Ctrl+S)'
  },
  menus: {
    move: 'Move',
    copy: 'Copy',
    compress: 'Compress',
    decompress: 'Extract to…'
  },
  recycle: {
    originalPath: 'Original Path',
    size: 'Size',
    deletedAt: 'Deleted At',
    alert: 'Deleted files go to the recycle bin first and can be restored or permanently deleted at any time; expired items are removed automatically by the "Recycle Bin Cleanup" scheduled task.'
  },
  dialogs: {
    mkdir: {
      title: 'New Folder',
      name: 'Folder Name',
      namePlaceholder: 'e.g. logs',
      path: 'Location'
    },
    mkfile: {
      title: 'New File',
      name: 'File Name',
      namePlaceholder: 'e.g. README.md',
      content: 'Initial Content (Optional)',
      contentPlaceholder: 'File content'
    },
    rename: {
      title: 'Rename',
      oldName: 'Current Name',
      newName: 'New Name'
    },
    move: {
      title: 'Move',
      source: 'Source Path',
      target: 'Target Directory',
      targetPlaceholder: '/path/to/dest'
    },
    copy: {
      title: 'Copy',
      source: 'Source Path',
      target: 'Target Directory',
      targetPlaceholder: '/path/to/dest'
    },
    compress: {
      title: 'Compress',
      source: 'Source Files',
      format: 'Archive Format',
      dest: 'Target File Path',
      destPlaceholder: '/path/to/archive.zip'
    },
    decompress: {
      title: 'Extract',
      archive: 'Archive',
      target: 'Target Directory',
      targetPlaceholder: '/path/to/extract'
    },
    upload: {
      title: 'Upload Files',
      target: 'Target Directory',
      dropTitle: 'Drag files here, or click to select',
      dropHint: 'Multiple files can be uploaded to the current directory at once'
    },
    delete: {
      title: 'Confirm Delete',
      content: 'Delete the selected {count} item(s)? This action cannot be undone.'
    },
    editor: {
      title: 'Editing: {name}',
      largeFile: 'Large file ({size}) — editing may be slow'
    }
  },
  messages: {
    loadDirFailed: 'Failed to load directory',
    searchFailed: 'Search failed',
    binaryDownload: 'Download binary files to view them',
    inputDirName: 'Please enter a folder name',
    dirCreated: 'Folder created',
    inputFileName: 'Please enter a file name',
    fileCreated: 'File created',
    createFailed: 'Creation failed',
    inputNewName: 'Please enter a new name',
    renameSuccess: 'Renamed successfully',
    renameFailed: 'Rename failed',
    moveSuccess: 'Moved successfully',
    moveFailed: 'Move failed',
    copySuccess: 'Copied successfully',
    copyFailed: 'Copy failed',
    inputTargetPath: 'Please enter a target path',
    compressSuccess: 'Compression task submitted',
    compressFailed: 'Compression failed',
    decompressSuccess: 'Extraction task submitted',
    decompressFailed: 'Extraction failed',
    deleteSuccess: 'Deleted successfully',
    deleteFailed: 'Delete failed',
    downloadSuccess: 'Download started',
    downloadFailed: 'Download failed',
    selectFileFirst: 'Please select files first',
    uploadFailed: 'Upload failed',
    uploadDone: 'Upload complete {done}/{total}',
    readFileFailed: 'Failed to read file',
    saveSuccess: 'Saved successfully',
    saveFailed: 'Save failed',
    loadRecycleFailed: 'Failed to load recycle bin',
    restored: 'Restored to original path',
    restoreFailed: 'Restore failed',
    purged: 'Permanently deleted'
  }
}
