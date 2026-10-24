// file 页面文案
export default {
  treeTitle: '目录树',
  recycleTitle: '回收站',
  buttons: {
    mkdir: '新建文件夹',
    mkfile: '新建文件',
    upload: '上传',
    refresh: '刷新'
  },
  search: {
    placeholder: '搜索当前目录…',
    searching: '搜索中',
    hint: '关键词「{keyword}」 · 命中 {count} 项',
    exit: '退出搜索'
  },
  batch: {
    selectedCount: '已选 {n} 项',
    copy: '复制',
    move: '移动',
    compress: '压缩',
    clear: '取消选择'
  },
  columns: {
    name: '名称',
    size: '大小',
    mode: '权限',
    modTime: '修改时间',
    actions: '操作'
  },
  actions: {
    rename: '重命名',
    download: '下载',
    delete: '删除',
    more: '更多',
    restore: '还原',
    purge: '彻底删除',
    remove: '移除',
    close: '关闭',
    save: '保存 (Ctrl+S)'
  },
  menus: {
    move: '移动',
    copy: '复制',
    compress: '压缩',
    decompress: '解压到…'
  },
  recycle: {
    originalPath: '原路径',
    size: '大小',
    deletedAt: '删除时间',
    alert: '删除的文件会先进入回收站，可随时还原或彻底删除；过期项由「回收站清理」计划任务自动清除。'
  },
  dialogs: {
    mkdir: {
      title: '新建文件夹',
      name: '文件夹名称',
      namePlaceholder: '例如：logs',
      path: '所在路径'
    },
    mkfile: {
      title: '新建文件',
      name: '文件名',
      namePlaceholder: '例如：README.md',
      content: '初始内容（可选）',
      contentPlaceholder: '文件内容'
    },
    rename: {
      title: '重命名',
      oldName: '原名称',
      newName: '新名称'
    },
    move: {
      title: '移动',
      source: '源路径',
      target: '目标目录',
      targetPlaceholder: '/path/to/dest'
    },
    copy: {
      title: '复制',
      source: '源路径',
      target: '目标目录',
      targetPlaceholder: '/path/to/dest'
    },
    compress: {
      title: '压缩',
      source: '源文件',
      format: '压缩格式',
      dest: '目标文件路径',
      destPlaceholder: '/path/to/archive.zip'
    },
    decompress: {
      title: '解压',
      archive: '压缩包',
      target: '目标目录',
      targetPlaceholder: '/path/to/extract'
    },
    upload: {
      title: '上传文件',
      target: '目标目录',
      dropTitle: '拖拽文件到此处，或点击选择',
      dropHint: '支持多文件同时上传到当前目录'
    },
    delete: {
      title: '确认删除',
      content: '确定删除选中的 {count} 项？此操作不可恢复。'
    },
    editor: {
      title: '编辑：{name}',
      largeFile: '文件较大（{size}），编辑可能较慢'
    }
  },
  messages: {
    loadDirFailed: '加载目录失败',
    searchFailed: '搜索失败',
    binaryDownload: '二进制文件请下载后查看',
    inputDirName: '请输入文件夹名称',
    dirCreated: '文件夹已创建',
    inputFileName: '请输入文件名',
    fileCreated: '文件已创建',
    createFailed: '创建失败',
    inputNewName: '请输入新名称',
    renameSuccess: '重命名成功',
    renameFailed: '重命名失败',
    moveSuccess: '移动成功',
    moveFailed: '移动失败',
    copySuccess: '复制成功',
    copyFailed: '复制失败',
    inputTargetPath: '请输入目标路径',
    compressSuccess: '压缩任务已提交',
    compressFailed: '压缩失败',
    decompressSuccess: '解压任务已提交',
    decompressFailed: '解压失败',
    deleteSuccess: '删除成功',
    deleteFailed: '删除失败',
    downloadSuccess: '开始下载',
    downloadFailed: '下载失败',
    selectFileFirst: '请先选择文件',
    uploadFailed: '上传失败',
    uploadDone: '上传完成 {done}/{total}',
    readFileFailed: '读取文件失败',
    saveSuccess: '保存成功',
    saveFailed: '保存失败',
    loadRecycleFailed: '加载回收站失败',
    restored: '已还原到原路径',
    restoreFailed: '还原失败',
    purged: '已彻底删除'
  }
}
