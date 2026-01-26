"use client"
import { BsGraphUp } from 'react-icons/bs';
import { MdAttachMoney } from 'react-icons/md';
import { Button } from "@workspace/ui/components/button";
import { useRouter } from 'next/navigation';
function Options() {
  const router = useRouter();
  return (
    <div className="flex items-center gap-2">
      <Button
        variant="ghost"
        className="gap-2 font-medium text-muted-foreground hover:text-foreground hover:bg-muted/50 transition-all cursor-pointer"
        onClick={() => router.push(process.env.NEXT_PUBLIC_BET_APP_URL ?? "")}
      >
        <MdAttachMoney className="text-lg" />
        Betting
      </Button>
      <Button
        variant="ghost"
        className="gap-2 font-medium text-muted-foreground hover:text-foreground hover:bg-muted/50 transition-all cursor-pointer"
        onClick={() => router.push(process.env.NEXT_PUBLIC_TRADE_APP_URL ?? "")}
      >
        <BsGraphUp className="text-lg" />
        Trading
      </Button>
    </div>
  )
}



export default Options;
