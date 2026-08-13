// Route /activity — lazy-load ActivityScreen component.
// Code-split màn hình Hoạt động, hiển thị skeleton loading khi chờ bundle.
import dynamic from "next/dynamic";
import RouteLoading from "@/components/app/RouteLoading";

const ActivityScreen = dynamic(() => import("@/components/screens/ActivityScreen"), {
  loading: () => <RouteLoading />,
});

export default function ActivityPage() {
  return <ActivityScreen />;
}
