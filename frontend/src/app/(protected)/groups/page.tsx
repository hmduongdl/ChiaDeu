import dynamic from "next/dynamic";
import RouteLoading from "@/components/app/RouteLoading";

const GroupsScreen = dynamic(() => import("@/components/screens/GroupsScreen"), {
  loading: () => <RouteLoading />,
});

export default function GroupsPage() {
  return <GroupsScreen />;
}
