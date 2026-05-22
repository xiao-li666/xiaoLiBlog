type HighlightToken = {
  start: number
  end: number
  className: string
}

type TokenRule = {
  className: string
  pattern: RegExp
}

const aliases: Record<string, string> = {
  cplusplus: 'cpp',
  cxx: 'cpp',
  hpp: 'cpp',
  js: 'javascript',
  jsx: 'javascript',
  mjs: 'javascript',
  ts: 'typescript',
  tsx: 'typescript',
  golang: 'go',
  py: 'python',
  rb: 'ruby',
  rs: 'rust',
  sh: 'bash',
  shell: 'bash',
  zsh: 'bash',
  yml: 'yaml',
  md: 'markdown',
  html: 'markup',
  xml: 'markup',
  svg: 'markup',
  vue: 'markup',
}

const keywordMap: Record<string, string[]> = {
  cpp: ['alignas', 'alignof', 'and', 'asm', 'auto', 'break', 'case', 'catch', 'class', 'concept', 'const', 'constexpr', 'continue', 'decltype', 'default', 'delete', 'do', 'else', 'enum', 'explicit', 'export', 'extern', 'for', 'friend', 'goto', 'if', 'inline', 'namespace', 'new', 'noexcept', 'operator', 'private', 'protected', 'public', 'requires', 'return', 'sizeof', 'static', 'struct', 'switch', 'template', 'this', 'throw', 'try', 'typedef', 'typename', 'union', 'using', 'virtual', 'volatile', 'while'],
  c: ['auto', 'break', 'case', 'const', 'continue', 'default', 'do', 'else', 'enum', 'extern', 'for', 'goto', 'if', 'inline', 'register', 'return', 'sizeof', 'static', 'struct', 'switch', 'typedef', 'union', 'volatile', 'while'],
  java: ['abstract', 'assert', 'break', 'case', 'catch', 'class', 'const', 'continue', 'default', 'do', 'else', 'enum', 'extends', 'final', 'finally', 'for', 'if', 'implements', 'import', 'instanceof', 'interface', 'new', 'package', 'private', 'protected', 'public', 'return', 'static', 'super', 'switch', 'synchronized', 'this', 'throw', 'throws', 'try', 'while'],
  javascript: ['async', 'await', 'break', 'case', 'catch', 'class', 'const', 'continue', 'debugger', 'default', 'delete', 'do', 'else', 'export', 'extends', 'finally', 'for', 'from', 'function', 'get', 'if', 'import', 'in', 'instanceof', 'let', 'new', 'of', 'return', 'set', 'static', 'super', 'switch', 'this', 'throw', 'try', 'typeof', 'var', 'while', 'yield'],
  typescript: ['abstract', 'as', 'async', 'await', 'break', 'case', 'catch', 'class', 'const', 'continue', 'declare', 'default', 'delete', 'do', 'else', 'enum', 'export', 'extends', 'finally', 'for', 'from', 'function', 'if', 'implements', 'import', 'in', 'infer', 'interface', 'is', 'keyof', 'let', 'module', 'namespace', 'new', 'of', 'private', 'protected', 'public', 'readonly', 'return', 'satisfies', 'static', 'super', 'switch', 'this', 'throw', 'try', 'type', 'typeof', 'var', 'while', 'yield'],
  go: ['break', 'case', 'chan', 'const', 'continue', 'default', 'defer', 'else', 'fallthrough', 'for', 'func', 'go', 'goto', 'if', 'import', 'interface', 'map', 'package', 'range', 'return', 'select', 'struct', 'switch', 'type', 'var'],
  python: ['and', 'as', 'assert', 'async', 'await', 'break', 'class', 'continue', 'def', 'del', 'elif', 'else', 'except', 'finally', 'for', 'from', 'global', 'if', 'import', 'in', 'is', 'lambda', 'nonlocal', 'not', 'or', 'pass', 'raise', 'return', 'try', 'while', 'with', 'yield'],
  rust: ['as', 'async', 'await', 'break', 'const', 'continue', 'crate', 'dyn', 'else', 'enum', 'extern', 'fn', 'for', 'if', 'impl', 'in', 'let', 'loop', 'match', 'mod', 'move', 'mut', 'pub', 'ref', 'return', 'self', 'static', 'struct', 'super', 'trait', 'type', 'unsafe', 'use', 'where', 'while'],
  php: ['abstract', 'and', 'array', 'as', 'break', 'case', 'catch', 'class', 'clone', 'const', 'continue', 'declare', 'default', 'do', 'echo', 'else', 'elseif', 'extends', 'final', 'finally', 'for', 'foreach', 'function', 'global', 'if', 'implements', 'include', 'interface', 'namespace', 'new', 'or', 'private', 'protected', 'public', 'require', 'return', 'static', 'switch', 'throw', 'trait', 'try', 'use', 'var', 'while'],
  sql: ['add', 'alter', 'and', 'as', 'asc', 'between', 'by', 'constraint', 'create', 'database', 'delete', 'desc', 'distinct', 'drop', 'exists', 'foreign', 'from', 'group', 'having', 'in', 'index', 'inner', 'insert', 'into', 'is', 'join', 'key', 'left', 'like', 'limit', 'not', 'null', 'on', 'or', 'order', 'outer', 'primary', 'references', 'right', 'select', 'set', 'table', 'union', 'unique', 'update', 'values', 'where'],
  bash: ['case', 'do', 'done', 'elif', 'else', 'esac', 'export', 'fi', 'for', 'function', 'if', 'in', 'local', 'read', 'select', 'then', 'until', 'while'],
  css: ['align-items', 'animation', 'background', 'border', 'box-shadow', 'color', 'display', 'flex', 'font-size', 'gap', 'grid', 'height', 'justify-content', 'line-height', 'margin', 'max-width', 'min-height', 'padding', 'position', 'transform', 'transition', 'width', 'z-index'],
}

const typeWords = ['any', 'bool', 'boolean', 'byte', 'char', 'double', 'float', 'int', 'int16', 'int32', 'int64', 'long', 'number', 'object', 'rune', 'short', 'string', 'uint', 'uint16', 'uint32', 'uint64', 'void']
const literalWords = ['false', 'None', 'nil', 'null', 'nullptr', 'true', 'undefined']

export function normalizeCodeLanguage(value = '') {
  const lang = value.trim().toLowerCase().split(/\s+/)[0]
  return aliases[lang] || lang
}

export function highlightCodeToHtml(source: string, language = '') {
  const tokens = highlightCodeTokens(source, language)
  if (!tokens.length) return escapeHtml(source)

  let html = ''
  let cursor = 0
  for (const token of tokens) {
    if (token.start < cursor) continue
    html += escapeHtml(source.slice(cursor, token.start))
    html += `<span class="${token.className}">${escapeHtml(source.slice(token.start, token.end))}</span>`
    cursor = token.end
  }
  html += escapeHtml(source.slice(cursor))
  return html
}

export function highlightCodeTokens(source: string, language = ''): HighlightToken[] {
  const lang = normalizeCodeLanguage(language)
  if (!source) return []
  if (lang === 'markup') return collectTokens(source, markupRules())
  if (lang === 'json') return collectTokens(source, jsonRules())
  if (lang === 'yaml') return collectTokens(source, yamlRules())
  if (lang === 'markdown') return collectTokens(source, markdownRules())
  if (lang === 'css') return collectTokens(source, cssRules())
  return collectTokens(source, genericRules(lang))
}

function collectTokens(source: string, rules: TokenRule[]) {
  const tokens: HighlightToken[] = []
  for (const rule of rules) {
    rule.pattern.lastIndex = 0
    let match: RegExpExecArray | null
    while ((match = rule.pattern.exec(source))) {
      const start = match.index
      const end = start + match[0].length
      if (end > start && !tokens.some((item) => rangesOverlap(start, end, item.start, item.end))) {
        tokens.push({ start, end, className: rule.className })
      }
      if (match[0].length === 0) rule.pattern.lastIndex += 1
    }
  }
  return tokens.sort((a, b) => a.start - b.start || b.end - a.end)
}

function genericRules(language: string): TokenRule[] {
  const keywords = keywordMap[language] || [...keywordMap.javascript, ...keywordMap.go, ...keywordMap.python]
  return [
    { className: 'hljs-comment', pattern: commentPattern(language) },
    { className: 'hljs-string', pattern: /"(?:\\.|[^"\\])*"|'(?:\\.|[^'\\])*'|`(?:\\.|[^`\\])*`/g },
    { className: 'hljs-number', pattern: /\b(?:0x[\da-fA-F]+|\d+(?:\.\d+)?(?:e[+-]?\d+)?)\b/g },
    { className: 'hljs-keyword', pattern: wordPattern(keywords) },
    { className: 'hljs-type', pattern: wordPattern(typeWords) },
    { className: 'hljs-literal', pattern: wordPattern(literalWords) },
    { className: 'hljs-title', pattern: /\b[A-Za-z_$][\w$]*(?=\s*\()/g },
    { className: 'hljs-variable', pattern: /\$[A-Za-z_][\w]*/g },
  ]
}

function markupRules(): TokenRule[] {
  return [
    { className: 'hljs-comment', pattern: /<!--[\s\S]*?-->/g },
    { className: 'hljs-meta', pattern: /<!doctype[^>]*>/gi },
    { className: 'hljs-tag', pattern: /<\/?|\/?>/g },
    { className: 'hljs-name', pattern: /(?<=<\/?)[A-Za-z][\w:-]*/g },
    { className: 'hljs-attr', pattern: /\s[A-Za-z_:][\w:.-]*(?=\s*=)/g },
    { className: 'hljs-string', pattern: /"(?:\\.|[^"\\])*"|'(?:\\.|[^'\\])*'/g },
  ]
}

function jsonRules(): TokenRule[] {
  return [
    { className: 'hljs-attr', pattern: /"(?:\\.|[^"\\])*"(?=\s*:)/g },
    { className: 'hljs-string', pattern: /"(?:\\.|[^"\\])*"/g },
    { className: 'hljs-number', pattern: /-?\b\d+(?:\.\d+)?(?:e[+-]?\d+)?\b/gi },
    { className: 'hljs-literal', pattern: /\b(?:false|null|true)\b/g },
  ]
}

function yamlRules(): TokenRule[] {
  return [
    { className: 'hljs-comment', pattern: /#.*/g },
    { className: 'hljs-attr', pattern: /^\s*[\w.-]+(?=\s*:)/gm },
    { className: 'hljs-string', pattern: /"(?:\\.|[^"\\])*"|'(?:\\.|[^'\\])*'/g },
    { className: 'hljs-number', pattern: /\b\d+(?:\.\d+)?\b/g },
    { className: 'hljs-literal', pattern: /\b(?:false|null|true)\b/g },
  ]
}

function markdownRules(): TokenRule[] {
  return [
    { className: 'hljs-meta', pattern: /^#{1,6}\s.+$/gm },
    { className: 'hljs-strong', pattern: /\*\*[^*]+\*\*/g },
    { className: 'hljs-string', pattern: /`[^`]+`/g },
    { className: 'hljs-link', pattern: /\[[^\]]+\]\([^)]+\)/g },
  ]
}

function cssRules(): TokenRule[] {
  return [
    { className: 'hljs-comment', pattern: /\/\*[\s\S]*?\*\//g },
    { className: 'hljs-string', pattern: /"(?:\\.|[^"\\])*"|'(?:\\.|[^'\\])*'/g },
    { className: 'hljs-selector-tag', pattern: /(^|[{};,])\s*[.#]?[A-Za-z][\w-]*(?=\s*[,{])/g },
    { className: 'hljs-attr', pattern: /[\w-]+(?=\s*:)/g },
    { className: 'hljs-number', pattern: /\b\d+(?:\.\d+)?(?:px|rem|em|vh|vw|%|s|ms)?\b/g },
    { className: 'hljs-built_in', pattern: /\b(?:auto|block|flex|grid|inline|none|relative|absolute|fixed|sticky|solid|transparent)\b/g },
    { className: 'hljs-keyword', pattern: wordPattern(keywordMap.css) },
  ]
}

function commentPattern(language: string) {
  if (language === 'python' || language === 'bash' || language === 'ruby') return /#.*/g
  if (language === 'sql') return /--.*|\/\*[\s\S]*?\*\//g
  return /\/\/.*|\/\*[\s\S]*?\*\//g
}

function wordPattern(words: string[]) {
  return new RegExp(`\\b(?:${words.map(escapeRegex).join('|')})\\b`, 'g')
}

function rangesOverlap(aStart: number, aEnd: number, bStart: number, bEnd: number) {
  return aStart < bEnd && bStart < aEnd
}

function escapeRegex(value: string) {
  return value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
}

function escapeHtml(value: string) {
  return value
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;')
}
