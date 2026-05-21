export type User = {
  id: number
  name: string
  email: string
  role: 'user' | 'admin'
  avatarUrl?: string
  createdAt?: string
}

export type Category = {
  id: number
  name: string
  slug: string
}

export type Article = {
  id: number
  title: string
  slug: string
  summary: string
  coverUrl: string
  content: string
  tags: string[]
  status: 'draft' | 'published'
  publishedAt?: string
  categoryId: number
  category?: Category
  author?: User
  likesCount: number
  favoritesCount: number
}

export type Comment = {
  id: number
  articleId: number
  body: string
  status: 'published' | 'hidden'
  createdAt: string
  user: User
}

export type AdminStats = {
  articles: {
    total: number
    published: number
    draft: number
  }
  categories: number
  comments: number
  users: number
  reactions: {
    likes: number
    favorites: number
  }
}
