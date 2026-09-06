// Component Modal tạo nhóm mới hoặc tham gia nhóm bằng mã chia sẻ.
// Cho phép chọn tab: "Tạo nhóm mới" hoặc "Tham gia bằng mã".
"use client";

import { useState } from "react";
import { apiFetch, ApiError } from "@/lib/api-client";
import { Group } from "@/types/api";

interface CreateGroupModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccess: (group: Group) => void;
}

export default function CreateGroupModal({ isOpen, onClose, onSuccess }: CreateGroupModalProps) {
  const [tab, setTab] = useState<"create" | "join">("create");
  const [name, setName] = useState("");
  const [currency, setCurrency] = useState("VND");
  const [shareCode, setShareCode] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  if (!isOpen) return null;

  // Xử lý tạo nhóm mới qua API POST /api/groups
  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) return;

    setIsLoading(true);
    setError(null);

    try {
      const res = await apiFetch<{ group: Group }>("/groups", {
        method: "POST",
        body: JSON.stringify({ name: name.trim(), currency }),
      });
      onSuccess(res.group);
      onClose();
      setName("");
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message);
      } else {
        setError("Không thể tạo nhóm. Vui lòng thử lại.");
      }
    } finally {
      setIsLoading(false);
    }
  };

  // Xử lý tham gia nhóm qua API POST /api/groups/join/:shareCode
  const handleJoin = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!shareCode.trim()) return;

    setIsLoading(true);
    setError(null);

    try {
      const res = await apiFetch<{ group: Group }>(`/groups/join/${shareCode.trim()}`, {
        method: "POST",
      });
      onSuccess(res.group);
      onClose();
      setShareCode("");
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message);
      } else {
        setError("Không thể tham gia nhóm. Vui lòng kiểm tra lại mã.");
      }
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4 backdrop-blur-sm">
      <div className="w-full max-w-sm rounded-2xl bg-white p-6 shadow-xl">
        {/* Tab chuyển đổi */}
        <div className="flex border-b border-gray-200 pb-3">
          <button
            type="button"
            onClick={() => { setTab("create"); setError(null); }}
            className={`flex-1 text-center font-semibold pb-2 border-b-2 transition-colors ${
              tab === "create"
                ? "border-[#047857] text-[#047857]"
                : "border-transparent text-gray-400"
            }`}
          >
            Tạo nhóm mới
          </button>
          <button
            type="button"
            onClick={() => { setTab("join"); setError(null); }}
            className={`flex-1 text-center font-semibold pb-2 border-b-2 transition-colors ${
              tab === "join"
                ? "border-[#047857] text-[#047857]"
                : "border-transparent text-gray-400"
            }`}
          >
            Tham gia bằng mã
          </button>
        </div>

        {error && (
          <div className="mt-3 rounded-lg bg-red-50 p-3 text-xs text-red-600">
            {error}
          </div>
        )}

        {tab === "create" ? (
          <form onSubmit={handleCreate} className="mt-4 space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700">Tên nhóm</label>
              <input
                type="text"
                required
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="VD: Chuyến đi Đà Lạt, Tiền nhà..."
                className="mt-1 w-full rounded-xl border border-gray-300 p-3 text-sm focus:border-[#047857] focus:outline-none"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700">Loại tiền tệ</label>
              <select
                value={currency}
                onChange={(e) => setCurrency(e.target.value)}
                className="mt-1 w-full rounded-xl border border-gray-300 p-3 text-sm focus:border-[#047857] focus:outline-none bg-white"
              >
                <option value="VND">VND (₫)</option>
                <option value="USD">USD ($)</option>
              </select>
            </div>

            <div className="flex gap-3 pt-2">
              <button
                type="button"
                onClick={onClose}
                className="flex-1 rounded-xl border border-gray-300 py-2.5 text-sm font-medium text-gray-700 hover:bg-gray-50"
              >
                Hủy
              </button>
              <button
                type="submit"
                disabled={isLoading}
                className="flex-1 rounded-xl bg-[#047857] py-2.5 text-sm font-medium text-white hover:bg-[#065f46] disabled:opacity-50"
              >
                {isLoading ? "Đang tạo..." : "Tạo nhóm"}
              </button>
            </div>
          </form>
        ) : (
          <form onSubmit={handleJoin} className="mt-4 space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700">Mã chia sẻ (Share Code)</label>
              <input
                type="text"
                required
                value={shareCode}
                onChange={(e) => setShareCode(e.target.value.toUpperCase())}
                placeholder="Nhập mã 6 ký tự..."
                maxLength={10}
                className="mt-1 w-full rounded-xl border border-gray-300 p-3 text-sm tracking-wider uppercase focus:border-[#047857] focus:outline-none"
              />
            </div>

            <div className="flex gap-3 pt-2">
              <button
                type="button"
                onClick={onClose}
                className="flex-1 rounded-xl border border-gray-300 py-2.5 text-sm font-medium text-gray-700 hover:bg-gray-50"
              >
                Hủy
              </button>
              <button
                type="submit"
                disabled={isLoading}
                className="flex-1 rounded-xl bg-[#047857] py-2.5 text-sm font-medium text-white hover:bg-[#065f46] disabled:opacity-50"
              >
                {isLoading ? "Đang xử lý..." : "Tham gia"}
              </button>
            </div>
          </form>
        )}
      </div>
    </div>
  );
}
