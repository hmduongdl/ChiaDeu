// Route /dashboard — lazy-load HomeScreen component.
// Sử dụng Next.js dynamic import để code-split màn hình Trang chủ,
// hiển thị RouteLoading skeleton trong khi chờ bundle tải về.
import dynamic from "next/dynamic";
import RouteLoading from "@/components/app/RouteLoading";

const HomeScreen = dynamic(() => import("@/components/screens/HomeScreen"), {
  loading: () => <RouteLoading />,
});

export default function DashboardPage() {
  return <HomeScreen />;
}
