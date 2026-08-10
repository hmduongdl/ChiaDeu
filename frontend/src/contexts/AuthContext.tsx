// ============================================================================
// AuthContext — Context quản lý trạng thái đăng nhập toàn app
// ============================================================================
// Sử dụng React Context API để chia sẻ user object giữa các component.
// AuthProvider bọc toàn bộ app trong layout.tsx.
// useAuth() hook để lấy user và setUser từ bất kỳ component nào.

'use client'

import { createContext, useContext, useState, ReactNode } from 'react'
import type { User } from '@/lib/types'

// Định nghĩa kiểu dữ liệu cho context
interface AuthContextType {
  user: User | null       // Người dùng hiện tại (null = chưa đăng nhập)
  setUser: (user: User | null) => void // Cập nhật user sau khi đăng nhập/đăng xuất
}

// Tạo context với giá trị mặc định
const AuthContext = createContext<AuthContextType>({
  user: null,
  setUser: () => {},
})

// AuthProvider: Component bọc toàn bộ app, cung cấp user state
export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)

  return (
    <AuthContext.Provider value={{ user, setUser }}>
      {children}
    </AuthContext.Provider>
  )
}

// useAuth: Hook tiện ích để truy cập AuthContext
export function useAuth() {
  return useContext(AuthContext)
}