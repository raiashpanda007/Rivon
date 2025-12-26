import { LoginCard } from "@/components/Auth/LoginCard"
import { Metadata } from "next"

export const metadata: Metadata = {
  title: "Login - Rivon",
}

function Page() {
  return (
    <div className="h-screen w-full flex items-center justify-center">
      <LoginCard />
    </div>
  )
}

export default Page
