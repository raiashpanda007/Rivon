import { useState, useEffect, useRef, RefObject } from 'react';

export function useScrollDirection(elementRef?: RefObject<HTMLElement | null>) {
    const [scrollDirection, setScrollDirection] = useState<"up" | "down" | null>(null);
    const lastScrollY = useRef(0);

    useEffect(() => {
        const target = elementRef?.current || window;

        const updateScrollDirection = () => {
            const scrollY = elementRef?.current ? elementRef.current.scrollTop : window.scrollY;
            const direction = scrollY > lastScrollY.current ? "down" : "up";

            if (direction !== scrollDirection && Math.abs(scrollY - lastScrollY.current) > 10) {
                setScrollDirection(direction);
            }
            lastScrollY.current = scrollY > 0 ? scrollY : 0;
        };

        target.addEventListener("scroll", updateScrollDirection);
        return () => target.removeEventListener("scroll", updateScrollDirection);
    }, [scrollDirection, elementRef]);

    return scrollDirection;
};
