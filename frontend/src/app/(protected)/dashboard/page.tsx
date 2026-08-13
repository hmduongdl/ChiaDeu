import dynamic from "next/dynamic";
import RouteLoading from "@/components/app/RouteLoading";

const HomeScreen = dynamic(() => import("@/components/screens/HomeScreen"), {
  loading: () => <RouteLoading />,
});

export default function DashboardPage() {
  return <HomeScreen />;
}
