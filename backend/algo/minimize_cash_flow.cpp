// ============================================================================
// minimize_cash_flow.cpp — Cài đặt thuật toán Minimize Cash Flow bằng C++
// ============================================================================
// Thuật toán: Greedy kết hợp Max-Heap (std::priority_queue)
// Mục tiêu: Tối thiểu hóa số lượng giao dịch thanh toán giữa các thành viên.
//
// Ý tưởng:
//   1. Tính net balance từng người: dương = creditor (được nhận), âm = debtor (phải trả)
//   2. Đưa creditors vào Max-Heap 1, debtors vào Max-Heap 2 (theo trị tuyệt đối)
//   3. Mỗi vòng lặp: lấy max creditor + max debtor, khớp khoản = min(|creditor|, |debtor|)
//   4. Cập nhật số dư, loại người có balance = 0 khỏi heap
//   5. Tiếp tục đến khi cả 2 heap rỗng → số settlement = tối thiểu
//
// Độ phức tạp: O(N log N)
//   - Tạo heap: O(N log N) với N = số thành viên
//   - Mỗi vòng lặp O(log N), tối đa N-1 vòng → O(N log N)
//   - So với cách tìm max tuyến tính O(N²) mỗi vòng → O(N²) tổng
//
// Ví dụ minh hoạ:
//   A: +100k (creditor)    B: -80k (debtor)    C: -20k (debtor)
//   Vòng 1: A(100k) vs B(80k) → B trả A 80k. A còn 20k, B hết.
//   Vòng 2: A(20k)  vs C(20k) → C trả A 20k. A hết, C hết.
//   Kết quả: 2 giao dịch (tối ưu, thay vì 3 nếu trả riêng lẻ)
// ============================================================================

#include "minimize_cash_flow.h"
#include <queue>      // std::priority_queue
#include <vector>     // std::vector
#include <string>     // std::string
#include <cstring>    // strdup, memset
#include <cmath>      // abs
#include <cstdlib>    // calloc, free
#include <algorithm>  // std::min

using namespace std;

// Person: cấu trúc lưu thông tin một người trong quá trình tính toán.
// id: UUID người dùng
// amount: số dư ròng (dương = creditor, âm = debtor)
struct Person {
    string id;
    double amount;
};

// MaxHeapComparator: so sánh 2 Person theo trị tuyệt đối của amount.
// Dùng cho priority_queue, ưu tiên người có |amount| lớn nhất ở đỉnh heap.
// Đây là Max-Heap vì priority_queue mặc định là Max-Heap (phần tử lớn nhất ra trước).
struct MaxHeapComparator {
    bool operator()(const Person& a, const Person& b) {
        // So sánh trị tuyệt đối: ai có |amount| lớn hơn thì đứng trước
        return abs(a.amount) < abs(b.amount);
    }
};

extern "C" {

// MinimizeCashFlow: Tính toán danh sách giao dịch thanh toán tối thiểu.
// Tham số:
//   userIDs: mảng các chuỗi UUID (C string)
//   balances: mảng số dư ròng tương ứng
//   n: số lượng phần tử
// Trả về: mảng Settlement kết thúc bằng sentinel (from = nullptr)
Settlement* MinimizeCashFlow(char** userIDs, double* balances, int n) {
    // Bước 1: Tách thành 2 Max-Heap: creditors (dương) và debtors (âm)
    // priority_queue với MaxHeapComparator: phần tử có |amount| lớn nhất ở đỉnh
    priority_queue<Person, vector<Person>, MaxHeapComparator> creditors;
    priority_queue<Person, vector<Person>, MaxHeapComparator> debtors;

    for (int i = 0; i < n; i++) {
        if (balances[i] > 0) {
            // Người có balance dương -> được nhận tiền -> đẩy vào heap creditors
            creditors.push({string(userIDs[i]), balances[i]});
        } else if (balances[i] < 0) {
            // Người có balance âm -> phải trả tiền -> đẩy vào heap debtors
            // Lưu ý: amount trong debtors là số âm
            debtors.push({string(userIDs[i]), balances[i]});
        }
        // balances[i] == 0: hoà vốn, không cần tham gia
    }

    // Cấp phát mảng kết quả: tối đa n*2 phần tử + 1 sentinel
    // Dùng calloc để tự động khởi tạo về 0 (sentinel)
    int maxSettlements = n * 2;
    Settlement* result = (Settlement*)calloc(maxSettlements + 1, sizeof(Settlement));
    int count = 0;

    // Bước 2: Lặp cho đến khi một trong 2 heap rỗng
    while (!creditors.empty() && !debtors.empty()) {
        // Lấy creditor có |amount| lớn nhất (người được nhận nhiều nhất)
        Person creditor = creditors.top();
        creditors.pop();

        // Lấy debtor có |amount| lớn nhất (người phải trả nhiều nhất)
        Person debtor = debtors.top();
        debtors.pop();

        // Số tiền khớp = min(số được nhận, số phải trả)
        // debtor.amount là số âm nên cần abs()
        double settleAmount = min(creditor.amount, abs(debtor.amount));

        // Tạo settlement: debtor trả cho creditor
        result[count].from = strdup(debtor.id.c_str());      // Người trả
        result[count].to = strdup(creditor.id.c_str());      // Người nhận
        result[count].amount = settleAmount;                  // Số tiền
        count++;

        // Cập nhật số dư sau khi khớp
        creditor.amount -= settleAmount;   // Creditor còn lại bao nhiêu?
        debtor.amount += settleAmount;     // Debtor còn nợ bao nhiêu? (amount âm + dương)

        // Đưa lại vào heap nếu còn dư (dùng epsilon 1e-9 để tránh sai số float)
        if (creditor.amount > 1e-9) {
            creditors.push(creditor);
        }
        if (abs(debtor.amount) > 1e-9) {
            debtors.push(debtor);
        }
    }

    // Đánh dấu sentinel: phần tử kết thúc mảng
    // calloc đã khởi tạo from = nullptr, to = nullptr, amount = 0
    // nên phần tử cuối cùng tự động là sentinel

    return result;
}

// FreeSettlementArray: Giải phóng toàn bộ bộ nhớ đã cấp phát cho mảng Settlement.
// Duyệt từng phần tử, free chuỗi from và to, sau đó free cả mảng.
// An toàn khi gọi với NULL (thoát sớm).
void FreeSettlementArray(Settlement* arr) {
    if (!arr) return; // Bảo vệ chống NULL pointer

    for (int i = 0; arr[i].from != nullptr; i++) {
        free(arr[i].from); // Giải phóng chuỗi from (được cấp bởi strdup)
        free(arr[i].to);   // Giải phóng chuỗi to (được cấp bởi strdup)
    }
    free(arr); // Giải phóng mảng
}

} // extern "C"