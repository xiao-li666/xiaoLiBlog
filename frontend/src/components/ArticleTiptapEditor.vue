<template>
  <section class="article-tiptap-editor">
    <div class="tiptap-toolbar" v-if="editor">
      <button
        type="button"
        class="tiptap-tool"
        :class="{ active: editor.isActive('heading', { level: 1 }) }"
        title="一级标题"
        @mousedown.prevent="editor.chain().focus().toggleHeading({ level: 1 }).run()"
      >
        <Heading1Icon :size="15" />
      </button>
      <button
        type="button"
        class="tiptap-tool"
        :class="{ active: editor.isActive('heading', { level: 2 }) }"
        title="二级标题"
        @mousedown.prevent="editor.chain().focus().toggleHeading({ level: 2 }).run()"
      >
        <Heading2Icon :size="15" />
      </button>
      <span class="tiptap-divider"></span>
      <button
        type="button"
        class="tiptap-tool"
        :class="{ active: editor.isActive('bold') }"
        title="加粗"
        @mousedown.prevent="editor.chain().focus().toggleBold().run()"
      >
        <BoldIcon :size="15" />
      </button>
      <button
        type="button"
        class="tiptap-tool"
        :class="{ active: editor.isActive('italic') }"
        title="斜体"
        @mousedown.prevent="editor.chain().focus().toggleItalic().run()"
      >
        <ItalicIcon :size="15" />
      </button>
      <button
        type="button"
        class="tiptap-tool"
        :class="{ active: editor.isActive('strike') }"
        title="删除线"
        @mousedown.prevent="editor.chain().focus().toggleStrike().run()"
      >
        <StrikethroughIcon :size="15" />
      </button>
      <button
        type="button"
        class="tiptap-tool"
        :class="{ active: editor.isActive('code') }"
        title="行内代码"
        @mousedown.prevent="editor.chain().focus().toggleCode().run()"
      >
        <CodeIcon :size="15" />
      </button>
      <span class="tiptap-divider"></span>
      <button
        type="button"
        class="tiptap-tool"
        :class="{ active: editor.isActive('bulletList') }"
        title="无序列表"
        @mousedown.prevent="editor.chain().focus().toggleBulletList().run()"
      >
        <ListIcon :size="15" />
      </button>
      <button
        type="button"
        class="tiptap-tool"
        :class="{ active: editor.isActive('orderedList') }"
        title="有序列表"
        @mousedown.prevent="editor.chain().focus().toggleOrderedList().run()"
      >
        <ListOrderedIcon :size="15" />
      </button>
      <button
        type="button"
        class="tiptap-tool"
        :class="{ active: editor.isActive('blockquote') }"
        title="引用"
        @mousedown.prevent="editor.chain().focus().toggleBlockquote().run()"
      >
        <QuoteIcon :size="15" />
      </button>
      <button
        type="button"
        class="tiptap-tool"
        :class="{ active: editor.isActive('codeBlock') }"
        title="代码块"
        @mousedown.prevent="toggleCodeBlock"
      >
        <Code2Icon :size="15" />
      </button>
      <select
        v-model="codeLanguage"
        class="tiptap-language-select"
        title="代码语言"
        @change="applyCodeLanguage"
      >
        <option value="">text</option>
        <option v-for="item in codeLanguages" :key="item.value" :value="item.value">
          {{ item.label }}
        </option>
      </select>
      <span class="tiptap-divider"></span>
      <button
        type="button"
        class="tiptap-tool"
        :class="{ active: editor.isActive('link') }"
        title="链接"
        @mousedown.prevent="setLink"
      >
        <LinkIcon :size="15" />
      </button>
      <button
        type="button"
        class="tiptap-tool"
        :class="{ uploading: imageUploading }"
        title="上传并插入图片"
        :disabled="imageUploading"
        @mousedown.prevent="openImagePicker"
      >
        <ImageIcon :size="15" />
      </button>
      <input
        ref="imageInputRef"
        class="tiptap-file-input"
        type="file"
        accept="image/png,image/jpeg,image/gif,image/webp"
        @change="handleImageUpload"
      />
      <button
        type="button"
        class="tiptap-tool"
        title="分割线"
        @mousedown.prevent="editor.chain().focus().setHorizontalRule().run()"
      >
        <MinusIcon :size="15" />
      </button>
      <span class="tiptap-spacer"></span>
      <span class="tiptap-count">{{ plainTextLength }} 字</span>
      <button
        type="button"
        class="tiptap-tool"
        title="撤销"
        :disabled="!editor.can().undo()"
        @mousedown.prevent="editor.chain().focus().undo().run()"
      >
        <Undo2Icon :size="15" />
      </button>
      <button
        type="button"
        class="tiptap-tool"
        title="重做"
        :disabled="!editor.can().redo()"
        @mousedown.prevent="editor.chain().focus().redo().run()"
      >
        <Redo2Icon :size="15" />
      </button>
    </div>

    <div class="tiptap-surface">
      <EditorContent :editor="editor" class="tiptap-content markdown-body" />
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { EditorContent, useEditor } from '@tiptap/vue-3'
import { Extension, mergeAttributes, Node } from '@tiptap/core'
import { Plugin, PluginKey, TextSelection } from '@tiptap/pm/state'
import { Decoration, DecorationSet } from '@tiptap/pm/view'
import StarterKit from '@tiptap/starter-kit'
import Link from '@tiptap/extension-link'
import Placeholder from '@tiptap/extension-placeholder'
import Typography from '@tiptap/extension-typography'
import { marked } from 'marked'
import TurndownService from 'turndown'
import {
  BoldIcon, Code2Icon, CodeIcon, Heading1Icon, Heading2Icon, ItalicIcon,
  ImageIcon, LinkIcon, ListIcon, ListOrderedIcon, MinusIcon, QuoteIcon, Redo2Icon,
  StrikethroughIcon, Undo2Icon,
} from 'lucide-vue-next'
import { api } from '../api'
import { highlightCodeTokens, normalizeCodeLanguage } from '../utils/codeHighlight'

const props = defineProps<{
  markdown: string
}>()

const codeLanguages = [
  { label: 'cpp', value: 'cpp' },
  { label: 'c', value: 'c' },
  { label: 'go', value: 'go' },
  { label: 'js', value: 'js' },
  { label: 'ts', value: 'ts' },
  { label: 'vue', value: 'vue' },
  { label: 'python', value: 'python' },
  { label: 'java', value: 'java' },
  { label: 'rust', value: 'rust' },
  { label: 'bash', value: 'bash' },
  { label: 'sql', value: 'sql' },
  { label: 'json', value: 'json' },
  { label: 'yaml', value: 'yaml' },
  { label: 'html', value: 'markup' },
  { label: 'css', value: 'css' },
]

const codeLanguage = ref('cpp')
const editorTick = ref(0)
const imageInputRef = ref<HTMLInputElement | null>(null)
const imageUploading = ref(false)

const ImageNode = Node.create({
  name: 'image',
  group: 'block',
  atom: true,
  draggable: true,

  addAttributes() {
    return {
      src: {
        default: null,
      },
      alt: {
        default: null,
      },
      title: {
        default: null,
      },
    }
  },

  parseHTML() {
    return [{ tag: 'img[src]' }]
  },

  renderHTML({ HTMLAttributes }) {
    return ['img', mergeAttributes(HTMLAttributes)]
  },
})

const SyntaxHighlightExtension = Extension.create({
  name: 'syntaxHighlight',
  addProseMirrorPlugins() {
    return [
      new Plugin({
        key: new PluginKey('syntaxHighlight'),
        props: {
          decorations(state) {
            const decorations: Decoration[] = []
            state.doc.descendants((node, pos) => {
              if (node.type.name !== 'codeBlock') return
              const language = normalizeCodeLanguage(String(node.attrs.language || ''))
              const text = node.textContent || ''
              const tokens = highlightCodeTokens(text, language)
              for (const token of tokens) {
                decorations.push(
                  Decoration.inline(pos + 1 + token.start, pos + 1 + token.end, {
                    class: token.className,
                  }),
                )
              }
            })
            return DecorationSet.create(state.doc, decorations)
          },
        },
      }),
    ]
  },
})

const turndown = new TurndownService({
  bulletListMarker: '-',
  codeBlockStyle: 'fenced',
  emDelimiter: '*',
  fence: '```',
  headingStyle: 'atx',
})

turndown.addRule('fencedCodeBlockWithLanguage', {
  filter(node) {
    return node.nodeName === 'PRE' && node.firstChild?.nodeName === 'CODE'
  },
  replacement(_content, node) {
    const code = node.firstChild as HTMLElement | null
    const className = code?.getAttribute('class') || ''
    const language = className.match(/language-([\w+-]+)/)?.[1] || ''
    const text = (code?.textContent || '').replace(/\n$/, '')
    return `\n\n\`\`\`${language}\n${text}\n\`\`\`\n\n`
  },
})

function markdownToHtml(source: string) {
  if (!source.trim()) return '<p></p>'
  return marked.parse(source, { async: false, breaks: false, gfm: true }) as string
}

function htmlToMarkdown(html: string) {
  return turndown.turndown(html).trim()
}

const editor = useEditor({
  content: markdownToHtml(props.markdown || ''),
  extensions: [
    StarterKit.configure({
      codeBlock: {
        enableTabIndentation: true,
        tabSize: 2,
      },
    }),
    Link.configure({
      autolink: true,
      linkOnPaste: true,
      openOnClick: false,
      HTMLAttributes: {
        rel: 'noopener noreferrer',
        target: '_blank',
      },
    }),
    Placeholder.configure({
      placeholder: '在这里直接写文章，Markdown 快捷语法会自动转换为排版效果',
    }),
    ImageNode,
    SyntaxHighlightExtension,
    Typography,
  ],
  editorProps: {
    attributes: {
      spellcheck: 'false',
      autocorrect: 'off',
      autocapitalize: 'off',
      translate: 'no',
    },
    handleKeyDown(view, event) {
      if (event.key !== 'Enter') return false
      const { state } = view
      const { selection, schema } = state
      if (!selection.empty) return false

      const { $from } = selection
      if ($from.parent.type.name !== 'paragraph') return false

      const match = $from.parent.textContent.trim().match(/^```([\w+-]+)?$|^~~~([\w+-]+)?$/)
      const codeBlockType = schema.nodes.codeBlock
      if (!match || !codeBlockType) return false

      event.preventDefault()
      const language = match[1] || match[2] || codeLanguage.value || null
      const tr = state.tr.replaceWith($from.before(), $from.after(), codeBlockType.create({ language }))
      tr.setSelection(TextSelection.create(tr.doc, $from.before() + 1))
      view.dispatch(tr)
      editorTick.value += 1
      return true
    },
  },
  onUpdate() {
    editorTick.value += 1
  },
  onSelectionUpdate({ editor }) {
    const language = editor.getAttributes('codeBlock').language
    if (language) codeLanguage.value = language
  },
})

const plainTextLength = computed(() => {
  editorTick.value
  return editor.value?.getText().trim().length ?? 0
})

watch(
  () => props.markdown,
  (markdown) => {
    if (!editor.value) return
    const currentMarkdown = htmlToMarkdown(editor.value.getHTML())
    if ((markdown || '').trim() === currentMarkdown.trim()) return
    editor.value.commands.setContent(markdownToHtml(markdown || ''), false)
    editorTick.value += 1
  },
)

function toggleCodeBlock() {
  if (!editor.value) return
  editor.value.chain().focus().toggleCodeBlock({ language: codeLanguage.value || undefined }).run()
}

function applyCodeLanguage() {
  if (!editor.value?.isActive('codeBlock')) return
  editor.value.chain().focus().updateAttributes('codeBlock', { language: codeLanguage.value || null }).run()
}

function setLink() {
  if (!editor.value) return
  const previousUrl = editor.value.getAttributes('link').href || ''
  const url = window.prompt('请输入链接地址', previousUrl)
  if (url === null) return

  if (!url.trim()) {
    editor.value.chain().focus().extendMarkRange('link').unsetLink().run()
    return
  }

  editor.value
    .chain()
    .focus()
    .extendMarkRange('link')
    .setLink({ href: normalizeUrl(url.trim()) })
    .run()
}

function openImagePicker() {
  if (imageUploading.value) return
  imageInputRef.value?.click()
}

async function handleImageUpload(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file || !editor.value) return

  imageUploading.value = true
  try {
    const res = await api.uploadAdminFile(file, 'image')
    if (!res.url) throw new Error('图片上传失败')
    editor.value
      .chain()
      .focus()
      .insertContent({
        type: 'image',
        attrs: {
          src: res.url,
          alt: file.name,
          title: file.name,
        },
      })
      .run()
    editorTick.value += 1
  } catch (error) {
    window.alert((error as Error).message || '图片上传失败')
  } finally {
    imageUploading.value = false
  }
}

function normalizeUrl(url: string) {
  if (/^(https?:|mailto:|tel:|\/)/i.test(url)) return url
  return `https://${url}`
}

function getMarkdown() {
  return editor.value ? htmlToMarkdown(editor.value.getHTML()) : props.markdown
}

function setMarkdown(markdown: string) {
  editor.value?.commands.setContent(markdownToHtml(markdown || ''), false)
  editorTick.value += 1
}

defineExpose({
  getMarkdown,
  setMarkdown,
})
</script>
