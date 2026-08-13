import dynamic from "next/dynamic";
import RouteLoading from "@/components/app/RouteLoading";

const ActivityScreen = dynamic(() => import("@/components/screens/ActivityScreen"), {
  loading: () => <RouteLoading />,
});

export default function ActivityPage() {
  return <ActivityScreen />;
}
