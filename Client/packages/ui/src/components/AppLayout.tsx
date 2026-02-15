"use client"

import { useRef, useState, useEffect } from "react";
import Header from "@workspace/ui/components/Header/Header";
import { ScrollArea } from "@workspace/ui/components/scroll-area";
import { useScrollDirection } from "@workspace/ui/hooks/useScrollDirection";
import { useDispatch, useSelector } from "react-redux";
import { RootState, setUserDetails, clearUserDetails, ProviderType } from "@workspace/store";
import ApiCaller, { RequestType } from "@workspace/api-caller";

interface AppLayoutProps {
    children: React.ReactNode;
    currentApp?: "trade" | "betting";
}

export function AppLayout({ children, currentApp }: AppLayoutProps) {
    const viewportRef = useRef<HTMLDivElement>(null);
    const scrollDirection = useScrollDirection(viewportRef);
    const [isScrolled, setIsScrolled] = useState(false);

    const dispatch = useDispatch();
    const userDetails = useSelector((state: RootState) => state.user.userDetails);

    useEffect(() => {
        const viewport = viewportRef.current;
        if (!viewport) return;

        const handleScroll = () => {
            setIsScrolled(viewport.scrollTop > 10);
        };

        viewport.addEventListener("scroll", handleScroll);
        return () => viewport.removeEventListener("scroll", handleScroll);
    }, []);

    useEffect(() => {
        if (userDetails) return;

        const checkAuth = async () => {
            try {
                const response = await ApiCaller<undefined, { id: string, type: string, name: string, email: string, provider: string, verified: boolean, profile: string }>({
                    requestType: RequestType.GET,
                    paths: ["api", "rivon", "auth", "me"],
                    body: undefined
                });

                if (response.ok && response.response.data) {
                    dispatch(setUserDetails({
                        id: response.response.data.id,
                        name: response.response.data.name,
                        email: response.response.data.email,
                        provider: response.response.data.provider as ProviderType,
                        verifiedStatus: response.response.data.verified,
                        profile: response.response.data.profile
                    }));
                }
            } catch (error) {
                // Not logged in
            }
        }
        checkAuth();
    }, [dispatch]);



    useEffect(() => {
        const handleStorage = (event: StorageEvent) => {
            if (event.key === 'logout-event') {
                dispatch(clearUserDetails());
                window.location.reload();
            }
        };
        window.addEventListener('storage', handleStorage);
        return () => window.removeEventListener('storage', handleStorage);
    }, [dispatch]);

    return (
        <div className="h-screen w-full flex flex-col overflow-hidden">
            <Header scrollDirection={scrollDirection} isScrolled={isScrolled} currentApp={currentApp} />
            <ScrollArea className="flex-1" viewportRef={viewportRef}>
                <div className="pt-16 min-h-full w-full">
                    {children}
                </div>
            </ScrollArea>
        </div>
    );
}
