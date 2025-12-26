"use client"
import { FcGoogle } from 'react-icons/fc';
import { FaGithub } from 'react-icons/fa';
import { Button } from "@workspace/ui/components/button"
import { useRouter } from 'next/navigation';
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

export function LoginCard() {
    const router = useRouter();
    return (
        <Card className="w-full max-w-sm bg-transparent">
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
            <CardContent>
                <form>
                    <div className="flex flex-col gap-6">
                        <div className="grid gap-2">
                            <Label className="font-heading font-semibold" htmlFor="email">Email</Label>
                            <Input
                                id="email"
                                type="email"
                                placeholder="ashwin@example.com"
                                required
                                className="placeholder:text-orange-500 placeholder:opacity-70 placeholder:font-semibold"
                            />
                        </div>
                        <div className="grid gap-2">
                            <div className="flex items-center">
                                <Label htmlFor="password">Password</Label>
                                <a href="#" className="ml-auto inline-block text-sm underline-offset-4 hover:underline">
                                    Forgot your password?
                                </a>
                            </div>
                            <Input id="password" type="password" required />
                        </div>
                    </div>
                </form>
            </CardContent>
            <CardFooter className="flex-col gap-2">
                <Button type="submit" className="w-full cursor-pointer font-heading font-semibold">
                    Sign In
                </Button>
                <Button variant="outline" className="w-full cursor-pointer">
                    <FcGoogle />Sign in with <span className='text-orange-500 font-semibold opacity-80'>Google</span>
                </Button>
                <Button variant="outline" className="w-full cursor-pointer">
                    <FaGithub />Sign in with <span className='text-orange-500 font-semibold opacity-80'>Github</span>
                </Button>
            </CardFooter>
        </Card>
    )
}
