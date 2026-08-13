// Component hiển thị chồng avatar của các thành viên trong nhóm.
// Props:
//   - avatars: mảng URL ảnh đại diện
//   - overflow: số thành viên còn lại (hiển thị dạng "+N")
//   - size: "small" (24px) hoặc "medium" (32px)
//   - muted: làm mờ và xám (grayscale) khi nhóm đã thanh toán xong
// Các avatar xếp chồng lên nhau với viền trắng và margin âm.
type AvatarStackProps = {
  avatars: readonly string[];
  overflow?: number;
  size?: "small" | "medium";
  muted?: boolean;
};

export default function AvatarStack({
  avatars,
  overflow,
  size = "small",
  muted = false,
}: AvatarStackProps) {
  const sizeClassName = size === "medium" ? "h-8 w-8" : "h-6 w-6";

  return (
    <div className={`flex items-center ${muted ? "grayscale opacity-60" : ""}`}>
      {avatars.map((avatar, index) => (
        <span
          key={`${avatar}-${index}`}
          className={`${sizeClassName} overflow-hidden rounded-full border-2 border-white ${
            index > 0 ? "-ml-2" : ""
          }`}
        >
          <img src={avatar} alt="" className="h-full w-full object-cover" />
        </span>
      ))}
      {overflow ? (
        <span
          className={`${sizeClassName} -ml-2 flex items-center justify-center rounded-full border-2 border-white bg-[#e2e8f8] text-[12px] font-semibold text-[#64748b]`}
        >
          +{overflow}
        </span>
      ) : null}
    </div>
  );
}
