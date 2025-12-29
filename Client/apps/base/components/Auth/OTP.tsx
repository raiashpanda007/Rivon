"use client"
const REGEXP_ONLY_DIGITS = "^[0-9]+$"
import ApiCaller, { RequestType } from "@workspace/api-caller";
import { Button } from "@workspace/ui/components/button"
import {
  InputOTP,
  InputOTPGroup,
  InputOTPSlot,
} from "@workspace/ui/components/input-otp"
import { ShowResponseToast } from "@workspace/ui/components/toast";
import { useEffect, useState } from "react";
import Loading from "../Loading";


async function RequestOTP() {
  const response = await ApiCaller<undefined, string>({ paths: ["api", "rivon", "auth", "verify", "send_otp"], requestType: RequestType.POST, retry: false });
  ShowResponseToast({
    heading: response.response.heading,
    message: response.response.message,
    statusCode: response.response.status,
    type: response.ok ? "INFORMATION" : "ERROR"

  })
}



export function InputOTPPattern() {
  const [error, setError] = useState("");
  const [value, setValue] = useState("");
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    const run = async () => {
      try {
        await RequestOTP();
      } catch (err) {
        console.error("Failed to request OTP", err);
      }
    };
    run();
  }, []);


  async function VerifyOTP(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setError("");

    if (value.length < 6) {
      setError("Please enter the complete 6-digit code");
      return;
    }
    const OTPStr = value.toString();
    setLoading(true);
    const response = await ApiCaller<{ otp: string }, string>({ requestType: RequestType.POST, body: { otp: OTPStr }, paths: ["api", "rivon", "auth", "verify", "verify_otp"], retry: false });
    console.log("Response :: ", response);
    setLoading(false);
    ShowResponseToast({
      heading: response.response.heading,
      message: response.response.message,
      statusCode: response.response.status,
      type: response.ok ? "SUCESS" : "ERROR"
    })


  }


  return (
    <>
      {loading && <Loading heading="Verifying account" message="Please wait while we verify your otp..." />}
      <form onSubmit={VerifyOTP} className="flex flex-col gap-4 justify-center items-center">
        <InputOTP maxLength={6} pattern={REGEXP_ONLY_DIGITS} value={value} onChange={(value) => setValue(value)}>
          <InputOTPGroup>
            <InputOTPSlot index={0} />
            <InputOTPSlot index={1} />
            <InputOTPSlot index={2} />
            <InputOTPSlot index={3} />
            <InputOTPSlot index={4} />
            <InputOTPSlot index={5} />
          </InputOTPGroup>
        </InputOTP>
        {error && <p className="text-red-500 text-sm font-medium animate-in fade-in slide-in-from-top-1">{error}</p>}
        <Button type="submit" className="bg-orange-500 hover:opacity-80 hover:bg-orange-500 cursor-pointer" >Submit </Button>
      </form>
    </>
  )
}
