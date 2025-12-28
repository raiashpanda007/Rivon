"use client"
import { motion } from "framer-motion";
import { FcGoogle } from 'react-icons/fc';
import { FaGithub } from 'react-icons/fa';
import { Button } from "@workspace/ui/components/button"
import { useRouter } from 'next/navigation';
import { useState } from 'react';
import {
  Card,
  CardAction,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@workspace/ui/components/card"
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

    if (!email || !password) {
      setError("Please fill in all fields");
      return;
    }

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
        verifiedStatus: response.response.data.verified
      }));

      if (!response.response.data.verified) {
        router.push("/auth/otp");
      }

      router.push("/dashboard"); // Or wherever they should go after login
    }
  };

  return (
    <>
      {isLoading && <Loading heading="Signing In" message="Please wait while we verify your credentials..." />}
      <motion.div
        initial={{ opacity: 0, y: 20, scale: 0.95 }}
        animate={{ opacity: 1, y: 0, scale: 1 }}
        transition={{ duration: 0.5, ease: [0.4, 0, 0.2, 1] }}
        className="w-full max-w-sm"
      >
        <Card className="bg-transparent">
          <CardHeader>
            <CardTitle>
              <svg xmlns="http://www.w3.org/2000/svg" width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="#f97316" strokeWidth="4" strokeLinecap="round" strokeLinejoin="round" className="drop-shadow-lg">
                <polyline points="23 6 13.5 15.5 8.5 10.5 1 18" />
                <polyline points="17 6 23 6 23 12" />
              </svg>
            </CardTitle>
            <CardDescription>
              Welcome back! Please sign in with your email or directly with your google or github account
            </CardDescription>

            <CardAction>
              <Button variant="default" className="bg-orange-500 dark:text-white font-semibold font-body cursor-pointer hover:opacity-80 hover:bg-orange-500 drop-shadow-md" onClick={() => router.push("/auth/register")}>Sign up</Button>
            </CardAction>
          </CardHeader>
          <form onSubmit={handleLogin}>
            <CardContent>
              <div className="flex flex-col gap-6">
                {error && (
                  <div className="p-3 text-sm text-red-500 bg-red-100 border border-red-200 rounded-md dark:bg-red-900/30 dark:border-red-900 dark:text-red-400">
                    {error}
                  </div>
                )}
                <div className="grid gap-2">
                  <Label className="font-heading font-semibold" htmlFor="email">Email</Label>
                  <Input
                    id="email"
                    type="email"
                    placeholder="ashwin@example.com"
                    required
                    className="placeholder:text-orange-500 placeholder:opacity-70 placeholder:font-semibold"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                  />
                </div>
                <div className="grid gap-2">
                  <div className="flex items-center">
                    <Label htmlFor="password">Password</Label>
                  </div>
                  <Input
                    id="password"
                    type="password"
                    required
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                  />
                </div>
              </div>
            </CardContent>
            <CardFooter className="flex-col gap-2">
              <Button type="submit" className="w-full cursor-pointer font-heading font-semibold">
                Sign In
              </Button>
              <Button variant="outline" type="button" className="w-full cursor-pointer" onClick={() => handleOAuthLogin('google')}>
                <FcGoogle />Sign in with <span className='text-orange-500 font-semibold opacity-80'>Google</span>
              </Button>
              <Button variant="outline" type="button" className="w-full cursor-pointer" onClick={() => handleOAuthLogin('github')}>
                <FaGithub />Sign in with <span className='text-orange-500 font-semibold opacity-80'>Github</span>
              </Button>
            </CardFooter>
          </form>
        </Card>
      </motion.div>
    </>
  )
}
