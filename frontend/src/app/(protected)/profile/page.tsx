// Route /profile — lazy-load ProfileScreen component.
// Code-split màn hình Cá nhân, hiển thị skeleton loading khi chờ bundle.
import dynamic from "next/dynamic";
import RouteLoading from "@/components/app/RouteLoading";

const ProfileScreen = dynamic(() => import("@/components/screens/ProfileScreen"), {
  loading: () => <RouteLoading />,
});

export default function ProfilePage() {
  return <ProfileScreen />;
}
