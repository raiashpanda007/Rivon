"use client"
import { Button } from "@workspace/ui/components/button";
import { useRouter } from "next/navigation.js";
function SignUp() {
  const router = useRouter();
  return (
    <Button onClick={() => router.push("/auth/register")} className="font-medium gap-2 bg-primary hover:bg-primary/90 text-primary-foreground shadow-sm cursor-pointer">
      Sign up
    </Button>
  )

}



export default SignUp;
