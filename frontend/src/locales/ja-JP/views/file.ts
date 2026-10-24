// file ページ文案
export default {
  treeTitle: 'ディレクトリツリー',
  recycleTitle: 'ゴミ箱',
  buttons: {
    mkdir: '新規フォルダ',
    mkfile: '新規ファイル',
    upload: 'アップロード',
    refresh: '更新'
  },
  search: {
    placeholder: '現在のディレクトリを検索…',
    searching: '検索中',
    hint: 'キーワード「{keyword}」 · {count} 件一致',
    exit: '検索を終了'
  },
  batch: {
    selectedCount: '{n} 件選択中',
    copy: 'コピー',
    move: '移動',
    compress: '圧縮',
    clear: '選択を解除'
  },
  columns: {
    name: '名前',
    size: 'サイズ',
    mode: '権限',
    modTime: '更新日時',
    actions: '操作'
  },
  actions: {
    rename: '名前を変更',
    download: 'ダウンロード',
    delete: '削除',
    more: 'その他',
    restore: '復元',
    purge: '完全削除',
    remove: '削除',
    close: '閉じる',
    save: '保存 (Ctrl+S)'
  },
  menus: {
    move: '移動',
    copy: 'コピー',
    compress: '圧縮',
    decompress: '展開…'
  },
  recycle: {
    originalPath: '元のパス',
    size: 'サイズ',
    deletedAt: '削除日時',
    alert: '削除されたファイルはまずゴミ箱に入り、いつでも復元または完全削除できます。期限切れの項目は「ゴミ箱のクリーンアップ」計画タスクにより自動的に削除されます。'
  },
  dialogs: {
    mkdir: {
      title: '新規フォルダ',
      name: 'フォルダ名',
      namePlaceholder: '例：logs',
      path: '作成先パス'
    },
    mkfile: {
      title: '新規ファイル',
      name: 'ファイル名',
      namePlaceholder: '例：README.md',
      content: '初期内容（任意）',
      contentPlaceholder: 'ファイル内容'
    },
    rename: {
      title: '名前を変更',
      oldName: '元の名前',
      newName: '新しい名前'
    },
    move: {
      title: '移動',
      source: 'ソースパス',
      target: '移動先ディレクトリ',
      targetPlaceholder: '/path/to/dest'
    },
    copy: {
      title: 'コピー',
      source: 'ソースパス',
      target: 'コピー先ディレクトリ',
      targetPlaceholder: '/path/to/dest'
    },
    compress: {
      title: '圧縮',
      source: 'ソースファイル',
      format: '圧縮形式',
      dest: '保存先ファイルパス',
      destPlaceholder: '/path/to/archive.zip'
    },
    decompress: {
      title: '展開',
      archive: '圧縮ファイル',
      target: '展開先ディレクトリ',
      targetPlaceholder: '/path/to/extract'
    },
    upload: {
      title: 'ファイルのアップロード',
      target: 'アップロード先ディレクトリ',
      dropTitle: 'ここにファイルをドラッグするか、クリックして選択してください',
      dropHint: '複数ファイルを同時に現在のディレクトリへアップロードできます'
    },
    delete: {
      title: '削除の確認',
      content: '選択した {count} 件を削除しますか？この操作は取り消せません。'
    },
    editor: {
      title: '編集：{name}',
      largeFile: 'ファイルが大きいため（{size}）、編集が遅くなる可能性があります'
    }
  },
  messages: {
    loadDirFailed: 'ディレクトリの読み込みに失敗しました',
    searchFailed: '検索に失敗しました',
    binaryDownload: 'バイナリファイルはダウンロードしてご覧ください',
    inputDirName: 'フォルダ名を入力してください',
    dirCreated: 'フォルダを作成しました',
    inputFileName: 'ファイル名を入力してください',
    fileCreated: 'ファイルを作成しました',
    createFailed: '作成に失敗しました',
    inputNewName: '新しい名前を入力してください',
    renameSuccess: '名前を変更しました',
    renameFailed: '名前の変更に失敗しました',
    moveSuccess: '移動しました',
    moveFailed: '移動に失敗しました',
    copySuccess: 'コピーしました',
    copyFailed: 'コピーに失敗しました',
    inputTargetPath: '宛先のパスを入力してください',
    compressSuccess: '圧縮タスクを送信しました',
    compressFailed: '圧縮に失敗しました',
    decompressSuccess: '展開タスクを送信しました',
    decompressFailed: '展開に失敗しました',
    deleteSuccess: '削除しました',
    deleteFailed: '削除に失敗しました',
    downloadSuccess: 'ダウンロードを開始しました',
    downloadFailed: 'ダウンロードに失敗しました',
    selectFileFirst: '先にファイルを選択してください',
    uploadFailed: 'アップロードに失敗しました',
    uploadDone: 'アップロード完了 {done}/{total}',
    readFileFailed: 'ファイルの読み込みに失敗しました',
    saveSuccess: '保存しました',
    saveFailed: '保存に失敗しました',
    loadRecycleFailed: 'ゴミ箱の読み込みに失敗しました',
    restored: '元のパスに復元しました',
    restoreFailed: '復元に失敗しました',
    purged: '完全に削除しました'
  }
}
