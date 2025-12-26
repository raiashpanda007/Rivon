import { RegisterCard } from "@/components/Auth/RegisterCard"
import { Metadata } from "next"

export const metadata: Metadata = {
  title: "Create Account - Rivon",
}
function Page() {
  return (
    <div className="h-screen w-full flex items-center justify-center">
      <RegisterCard />
    </div>
  )
}

export default Page
