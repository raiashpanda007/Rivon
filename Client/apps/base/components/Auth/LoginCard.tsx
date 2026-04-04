"use client"
import { motion } from "framer-motion";
import { FcGoogle } from 'react-icons/fc';
import { FaGithub } from 'react-icons/fa';
import { Button } from "@workspace/ui/components/button"
import { useRouter } from 'next/navigation';
import { useState } from 'react';
import { Input } from "@workspace/ui/components/input"
import { Label } from "@workspace/ui/components/label"
import ApiCaller, { RequestType, getOAuthUrl } from "@workspace/api-caller"
import { useDispatch } from "react-redux";
import { setUserDetails, ProviderType } from "@workspace/store";
import { ShowResponseToast } from "@workspace/ui/components/toast";
import Loading from "../Loading";

export function LoginCard() {
  const router = useRouter();
  const dispatch = useDispatch();
  const [email, setEmail] = useState<string>("");
  const [password, setPassword] = useState<string>("");
  const [error, setError] = useState<string>("");
  const [isLoading, setIsLoading] = useState<boolean>(false);

  const handleOAuthLogin = (provider: 'google' | 'github') => {
    window.location.href = getOAuthUrl(provider);
  }

  async function handleLogin(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    if (!email || !password) { setError("Please fill in all fields"); return; }
    setIsLoading(true);
    const response = await ApiCaller<{ email: string, password: string }, { id: string, type: string, name: string, email: string, provider: string, verified: boolean }>({
      requestType: RequestType.POST,
      paths: ["api", "rivon", "auth", "credentials", "signin"],
      body: { email, password },
      retry: false
    })
    setIsLoading(false);
    ShowResponseToast({
      heading: response.response.heading,
      message: response.response.message,
      statusCode: response.response.status,
      type: response.ok ? "SUCESS" : "ERROR"
    })
    if (response.ok) {
      dispatch(setUserDetails({
        id: response.response.data.id,
        name: response.response.data.name,
        email: response.response.data.email,
        provider: response.response.data.provider as ProviderType,
        verifiedStatus: response.response.data.verified,
        profile: ""
      }));
      if (!response.response.data.verified) router.push("/auth/otp");
      router.push("/dashboard");
    }
  };

  return (
    <>
      {isLoading && <Loading heading="Signing In" message="Please wait while we verify your credentials..." />}
      <motion.div
        initial={{ opacity: 0, y: 16, scale: 0.97 }}
        animate={{ opacity: 1, y: 0, scale: 1 }}
        transition={{ duration: 0.4, ease: [0.4, 0, 0.2, 1] }}
        className="w-full"
      >
        {/* Terminal panel */}
        <div className="terminal-panel overflow-hidden">
          {/* Panel header */}
          <div className="terminal-panel-header">
            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#f97316" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round">
              <polyline points="23 6 13.5 15.5 8.5 10.5 1 18" />
              <polyline points="17 6 23 6 23 12" />
            </svg>
            <span className="font-mono text-2xl text-muted-foreground"> SIGN_IN</span>
            <Button
              variant="ghost"
              size="sm"
              className="ml-auto h-6 px-2 text-[10px] font-mono text-orange-400 hover:text-orange-500 hover:bg-orange-500/10 cursor-pointer"
              onClick={() => router.push("/auth/register")}
            >
              REGISTER →
            </Button>
          </div>

          <div className="px-4 py-3">
            <p className="font-mono text-[10px] text-muted-foreground mb-4">
              Welcome back. Authenticate to access your trading terminal.
            </p>

            {error && (
              <div className="mb-4 px-3 py-2 border border-red-500/30 bg-red-500/10 rounded-sm">
                <p className="font-mono text-[10px] text-red-400">{error}</p>
              </div>
            )}

            <form onSubmit={handleLogin} className="space-y-4">
              <div className="space-y-1.5">
                <Label htmlFor="email" className="font-mono text-[10px] text-muted-foreground">
                  EMAIL
                </Label>
                <Input
                  id="email"
                  type="email"
                  placeholder="user@domain.com"
                  required
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  className="font-mono text-xs h-9 bg-input border-border focus:border-orange-500/50 focus:ring-orange-500/20 rounded-sm"
                />
              </div>

              <div className="space-y-1.5">
                <Label htmlFor="password" className="font-mono text-[10px] text-muted-foreground">
                  PASSWORD
                </Label>
                <Input
                  id="password"
                  type="password"
                  required
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  className="font-mono text-xs h-9 bg-input border-border focus:border-orange-500/50 focus:ring-orange-500/20 rounded-sm"
                />
              </div>

              <Button
                type="submit"
                className="w-full h-9 bg-orange-500 hover:bg-orange-600 text-white border-0 rounded-sm font-mono text-xs tracking-wide cursor-pointer shadow-[0_0_16px_rgba(249,115,22,0.2)]"
              >
                AUTHENTICATE
              </Button>
            </form>

            <div className="flex items-center gap-3 my-4">
              <div className="h-px flex-1 bg-border" />
              <span className="font-mono text-[9px] text-muted-foreground/60">OR CONTINUE WITH</span>
              <div className="h-px flex-1 bg-border" />
            </div>

            <div className="grid grid-cols-2 gap-2">
              <Button
                variant="outline"
                type="button"
                className="h-9 border-border rounded-sm font-mono text-xs cursor-pointer hover:border-orange-500/40"
                onClick={() => handleOAuthLogin('google')}
              >
                <FcGoogle className="mr-1.5" />
                Google
              </Button>
              <Button
                variant="outline"
                type="button"
                className="h-9 border-border rounded-sm font-mono text-xs cursor-pointer hover:border-orange-500/40"
                onClick={() => handleOAuthLogin('github')}
              >
                <FaGithub className="mr-1.5" />
                GitHub
              </Button>
            </div>
          </div>
        </div>
      </motion.div>
    </>
  )
}
