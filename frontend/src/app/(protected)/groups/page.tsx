// Route /groups — lazy-load GroupsScreen component.
// Code-split màn hình Nhóm, hiển thị skeleton loading khi chờ bundle.
import dynamic from "next/dynamic";
import RouteLoading from "@/components/app/RouteLoading";

const GroupsScreen = dynamic(() => import("@/components/screens/GroupsScreen"), {
  loading: () => <RouteLoading />,
});

export default function GroupsPage() {
  return <GroupsScreen />;
}
