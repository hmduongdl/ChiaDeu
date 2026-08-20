// Types định nghĩa các thực thể API được đồng bộ từ Go models (backend/models)
// QUY TẮC: Bất kỳ thay đổi nào trong struct Go phải được cập nhật đồng bộ tại đây.

export interface User {
  id: string;
  name: string;
  email: string;
  phone?: string;
  avatarUrl?: string;
  createdAt: string; // Định dạng ISO Date String từ time.Time
  updatedAt?: string;
}

export type GroupRole = "ADMIN" | "MEMBER";
export type MemberStatus = "ACTIVE" | "LEFT";
export type GroupStatus = "ACTIVE" | "ARCHIVED";

export interface Group {
  id: string;
  name: string;
  createdBy: string;
  shareCode: string;
  currency: string;
  status: GroupStatus;
  createdAt: string;
  updatedAt?: string;
}

export interface GroupMember {
  groupId: string;
  userId: string;
  role: GroupRole;
  status: MemberStatus;
  joinedAt: string;
  leftAt?: string;
}

export type SplitType = "EQUAL" | "PERCENT" | "WEIGHT" | "CUSTOM";
export type ExpenseStatus = "ACTIVE" | "VOIDED";

export interface Expense {
  id: string;
  groupId: string;
  createdBy: string;
  paidBy: string;
  description: string;
  amountMinor: number; // Go int64 ánh xạ thành number trong TS
  splitType: SplitType;
  expenseDate: string;
  batchId?: string;
  status: ExpenseStatus;
  createdAt: string;
  updatedAt?: string;
}

export interface ExpenseSplit {
  id: string;
  expenseId: string;
  userId: string;
  shareMinor: number;
  createdAt: string;
}

export type BatchStatus = "OPEN" | "COMPLETED" | "CANCELLED";
export type SettlementStatus = "PENDING" | "AWAITING_CONFIRMATION" | "PAID" | "CANCELLED";

export interface SettlementBatch {
  id: string;
  groupId: string;
  createdBy: string;
  idempotencyKey: string;
  status: BatchStatus;
  createdAt: string;
  completedAt?: string;
  cancelledAt?: string;
}

export interface Settlement {
  id: string;
  batchId: string;
  fromUserId: string;
  toUserId: string;
  amountMinor: number;
  paymentCode: string;
  status: SettlementStatus;
  markedSentAt?: string;
  paidAt?: string;
  createdAt: string;
  updatedAt?: string;
}

export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  error?: string;
}
