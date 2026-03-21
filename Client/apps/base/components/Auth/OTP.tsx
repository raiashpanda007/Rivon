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
  const response = await ApiCaller<undefined, string>({ paths: ["api", "rivon", "auth", "verify", "send_otp"], requestType: RequestType.POST });
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
    const response = await ApiCaller<{ otp: string }, string>({ requestType: RequestType.POST, body: { otp: OTPStr }, paths: ["api", "rivon", "auth", "verify", "verify_otp"] });
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
      <div className="w-full max-w-md mx-auto">
        <div className="terminal-panel overflow-hidden">
          <div className="terminal-panel-header">
            <span className="dot-live" />
            <span className="font-mono text-[10px] text-muted-foreground ml-2 tracking-widest">VERIFICATION_REQUIRED</span>
          </div>
          <div className="px-6 py-8 flex flex-col items-center">
            <p className="font-mono text-xs text-muted-foreground mb-8 text-center max-w-xs">
              Enter the 6-digit authorization code sent to your device to secure your terminal.
            </p>
            <form onSubmit={VerifyOTP} className="flex flex-col gap-6 w-full items-center">
              <InputOTP maxLength={6} pattern={REGEXP_ONLY_DIGITS} value={value} onChange={(value) => setValue(value)}>
                <InputOTPGroup>
                  <InputOTPSlot index={0} className="border-border rounded-sm h-12 w-10 font-mono text-lg" />
                  <InputOTPSlot index={1} className="border-border rounded-sm h-12 w-10 font-mono text-lg" />
                  <InputOTPSlot index={2} className="border-border rounded-sm h-12 w-10 font-mono text-lg" />
                  <InputOTPSlot index={3} className="border-border rounded-sm h-12 w-10 font-mono text-lg" />
                  <InputOTPSlot index={4} className="border-border rounded-sm h-12 w-10 font-mono text-lg" />
                  <InputOTPSlot index={5} className="border-border rounded-sm h-12 w-10 font-mono text-lg" />
                </InputOTPGroup>
              </InputOTP>
              {error && <p className="text-red-400 font-mono text-xs animate-in fade-in slide-in-from-top-1 bg-red-500/10 px-3 py-1.5 rounded-sm">{error}</p>}
              <Button type="submit" className="w-full max-w-xs bg-orange-500 hover:opacity-80 hover:bg-orange-500 cursor-pointer text-white font-mono text-xs rounded-sm shadow-[0_0_16px_rgba(249,115,22,0.2)]">VERIFY_CODE</Button>
            </form>
          </div>
        </div>
      </div>
    </>
  )
}
