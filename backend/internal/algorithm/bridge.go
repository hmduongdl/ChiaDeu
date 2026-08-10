// Package algorithm là cầu nối (bridge) giữa Go backend và thư viện C++.
// Sử dụng cgo để gọi hàm MinimizeCashFlow từ shared library (.so) được biên dịch từ C++.
// Thuật toán: Greedy với Max-Heap, độ phức tạp O(N log N).
//
// Cách build:
//  1. cd backend/algo && make        # Biên dịch C++ → libminimizecashflow.so
//  2. go build                           # Go tự động link với .so qua cgo
package algorithm

/*
#cgo LDFLAGS: -L${SRCDIR}/../../algo -lminimizecashflow
#include <stdlib.h>
#include "minimize_cash_flow.h"
*/
import "C"
import (
	"fmt"
	"unsafe"
)

// Balance thể hiện số dư ròng (net balance) của một người dùng.
// Amount > 0: người này được nhận tiền (creditor - chủ nợ)
// Amount < 0: người này phải trả tiền (debtor - con nợ)
// Amount = 0: hoà vốn, không cần tham gia thanh toán
type Balance struct {
	UserID string  // UUID của người dùng
	Amount float64 // Số dư ròng (dương = được nhận, âm = phải trả)
}

// Settlement thể hiện một giao dịch thanh toán tối ưu giữa 2 người.
// From: người phải trả tiền (debtor)
// To: người được nhận tiền (creditor)
// Amount: số tiền cần thanh toán
type Settlement struct {
	From   string  // UUID người trả
	To     string  // UUID người nhận
	Amount float64 // Số tiền
}

// MinimizeCashFlow tính toán danh sách giao dịch thanh toán tối thiểu.
// Đầu vào: danh sách Balance của từng thành viên trong nhóm.
// Đầu ra: danh sách Settlement - mỗi settlement là một giao dịch cần thực hiện.
//
// Thuật toán (cài đặt bên C++):
//  1. Tách thành 2 Max-Heap: creditors (Amount > 0) và debtors (Amount < 0)
//  2. Heap sắp xếp theo trị tuyệt đối của Amount giảm dần
//  3. Lặp: lấy max creditor + max debtor, tạo settlement = min(|creditor|, |debtor|)
//  4. Cập nhật số dư, loại người có balance = 0 khỏi heap
//  5. Tiếp tục đến khi cả 2 heap rỗng
//  6. Độ phức tạp: O(N log N) nhờ priority_queue thay vì tìm max tuyến tính
func MinimizeCashFlow(balances []Balance) ([]Settlement, error) {
	n := len(balances)
	if n == 0 {
		return nil, nil
	}

	// Chuyển đổi dữ liệu Go sang C types để truyền qua cgo
	cUserIDs := make([]*C.char, n) // Mảng con trỏ char* cho user IDs
	cAmounts := make([]C.double, n) // Mảng double cho số dư

	for i, b := range balances {
		cUserIDs[i] = C.CString(b.UserID) // C.CString cấp phát bộ nhớ trong C
		defer C.free(unsafe.Pointer(cUserIDs[i])) // Giải phóng sau khi hàm kết thúc
		cAmounts[i] = C.double(b.Amount)
	}

	// Gọi hàm C++: MinimizeCashFlow(char** userIDs, double* balances, int n)
	// Trả về mảng Settlement* kết thúc bằng sentinel (from = NULL)
	result := C.MinimizeCashFlow(&cUserIDs[0], &cAmounts[0], C.int(n))

	if result == nil {
		return nil, fmt.Errorf("thuật toán trả về null")
	}

	// Chuyển đổi kết quả từ C về Go
	settlements := make([]Settlement, 0)
	for i := 0; result[i].from != nil; i++ {
		settlements = append(settlements, Settlement{
			From:   C.GoString(result[i].from), // Chuyển char* → Go string
			To:     C.GoString(result[i].to),
			Amount: float64(result[i].amount),
		})
	}

	// Giải phóng bộ nhớ bên C (mảng và các chuỗi bên trong)
	C.FreeSettlementArray(result)

	return settlements, nil
}