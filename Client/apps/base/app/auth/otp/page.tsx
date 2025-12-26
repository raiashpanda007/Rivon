import { InputOTPPattern } from "@/components/Auth/OTP"
import { Button } from "@workspace/ui/components/button";

function Page() {
  return (
    <div className="h-screen w-full flex flex-col items-center justify-center gap-4">
      <h1 className="text-2xl font-bold">Enter <span className="text-orange-500" children="OTP" /> </h1>
      <p className="text-muted-foreground text-sm mb-4">Please enter the 6-digit code sent to your email.</p>
      <InputOTPPattern />
      <Button className="bg-orange-500 hover:opacity-80 hover:bg-orange-500 cursor-pointer" >Submit </Button>
    </div>
  )
}


export default Page;
