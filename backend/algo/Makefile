# ============================================================================
# Makefile — Biên dịch thư viện C++ Minimize Cash Flow thành shared library (.so)
# ============================================================================
# Thư viện này được Go backend gọi thông qua cgo.
# Yêu cầu: g++ (hỗ trợ C++17) và hệ điều hành Linux.
#
# Cách dùng:
#   make          # Biên dịch → libminimizecashflow.so
#   make clean    # Xoá file .so
#
# Cờ biên dịch:
#   -std=c++17:    Sử dụng chuẩn C++17 (cho std::priority_queue, structured bindings...)
#   -O2:           Tối ưu hoá level 2 (cân bằng giữa tốc độ và kích thước)
#   -fPIC:         Position Independent Code (bắt buộc cho shared library)
#   -shared:       Tạo shared library (.so) thay vì executable
# ============================================================================

CXX := g++
CXXFLAGS := -std=c++17 -O2 -fPIC
TARGET := libminimizecashflow.so

# Mục tiêu mặc định: biên dịch file .cpp thành shared library
$(TARGET): minimize_cash_flow.cpp minimize_cash_flow.h
	$(CXX) $(CXXFLAGS) -shared -o $(TARGET) minimize_cash_flow.cpp

# Dọn dẹp file biên dịch
clean:
	rm -f $(TARGET)

.PHONY: clean