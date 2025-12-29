import { InputOTPPattern } from "@/components/Auth/OTP"
import { Metadata } from "next"

export const metadata: Metadata = {
  title: "Verify Account - Rivon",
}
function Page() {
  return (
    <div className="h-screen w-full flex flex-col items-center justify-center gap-4">
      <h1 className="text-2xl font-bold">Enter <span className="text-orange-500" children="OTP" /> </h1>
      <p className="text-muted-foreground text-sm mb-4">Please enter the 6-digit code sent to your email.</p>
      <InputOTPPattern />

    </div>
  )
}


export default Page;
