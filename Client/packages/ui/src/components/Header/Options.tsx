"use client"
import { BsGraphUp } from 'react-icons/bs';
import { Button } from "@workspace/ui/components/button";
import { useRouter } from 'next/navigation';
interface OptionsProps {
  currentApp?: "trade";
}

function Options({ currentApp }: OptionsProps) {
  const router = useRouter();
  return (
    <div className="flex items-center gap-2">
      <Button
        variant="ghost"
        className={`gap-2 font-medium transition-all cursor-pointer ${currentApp === "trade"
            ? "bg-muted text-foreground"
            : "text-muted-foreground hover:text-foreground hover:bg-muted/50"
          }`}
        onClick={() => router.push(process.env.NEXT_PUBLIC_TRADE_APP_URL ?? "")}
      >
        <BsGraphUp className="text-lg" />
        Trading
      </Button>
    </div>
  )
}



export default Options;
