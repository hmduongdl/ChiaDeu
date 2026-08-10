// ============================================================================
// minimize_cash_flow.h — Header C cho thuật toán Minimize Cash Flow
// ============================================================================
// File này định nghĩa interface C-compatible để Go backend có thể gọi
// thư viện C++ qua cgo. Sử dụng extern "C" để tránh name mangling của C++.
//
// Cấu trúc dữ liệu:
//   - Settlement: một giao dịch thanh toán giữa 2 người (from -> to, amount)
//   - Mảng Settlement kết thúc bằng sentinel: phần tử có from == NULL
//
// Cách dùng:
//   1. Gọi MinimizeCashFlow(userIDs, balances, n) để tính toán
//   2. Duyệt mảng kết quả đến khi gặp sentinel (from == NULL)
//   3. Gọi FreeSettlementArray() để giải phóng bộ nhớ
// ============================================================================

#ifndef MINIMIZE_CASH_FLOW_H
#define MINIMIZE_CASH_FLOW_H

#ifdef __cplusplus
extern "C" {
#endif

// Settlement: một giao dịch thanh toán tối ưu giữa 2 người.
// from: người phải trả tiền (debtor)
// to: người được nhận tiền (creditor)
// amount: số tiền cần thanh toán
typedef struct {
    char* from;    // Chuỗi UUID người trả (được cấp phát bởi strdup)
    char* to;      // Chuỗi UUID người nhận (được cấp phát bởi strdup)
    double amount; // Số tiền
} Settlement;

// MinimizeCashFlow: Tính toán danh sách giao dịch thanh toán tối thiểu.
// Tham số:
//   userIDs: mảng các chuỗi UUID của người dùng (kích thước n)
//   balances: mảng số dư ròng tương ứng (dương = được nhận, âm = phải trả)
//   n: số lượng người dùng
// Trả về:
//   Mảng Settlement kết thúc bằng sentinel (from = NULL).
//   Người gọi PHẢI gọi FreeSettlementArray() để giải phóng bộ nhớ.
// Độ phức tạp: O(N log N) với N = số người dùng.
Settlement* MinimizeCashFlow(char** userIDs, double* balances, int n);

// FreeSettlementArray: Giải phóng bộ nhớ của mảng Settlement.
// Giải phóng tất cả chuỗi from, to bên trong từng phần tử và cả mảng.
// An toàn khi gọi với NULL.
void FreeSettlementArray(Settlement* arr);

#ifdef __cplusplus
}
#endif

#endif // MINIMIZE_CASH_FLOW_H