"use client"

import { toast } from "sonner"
import { CheckCircle, XCircle, AlertCircle } from "lucide-react"
import { cn } from "../lib/utils"

interface ShowResponseToastProps {
  heading: string;
  message: string;
  statusCode: number;
  type: "ERROR" | "SUCESS" | "INFORMATION"
}

export function ShowResponseToast({ heading, message, statusCode, type }: ShowResponseToastProps) {
  toast.custom(() => (
    <div
      className={cn(
        "relative flex w-full max-w-xl gap-4 overflow-hidden rounded-lg border bg-background p-4 shadow-lg transition-all",
        "dark:bg-zinc-950 dark:border-zinc-800"
      )}
    >
      {/* Right Accent Border */}
      <div
        className={cn(
          "absolute right-0 top-0 bottom-0 w-1.5",
          type === "SUCESS" && "bg-green-500",
          type === "ERROR" && "bg-red-500",
          type === "INFORMATION" && "bg-yellow-500"
        )}
      />

      {/* Icon */}
      <div className="flex shrink-0 items-start pt-0.5">
        {type === "SUCESS" && <CheckCircle className="size-6 text-green-500" />}
        {type === "ERROR" && <XCircle className="size-6 text-red-500" />}
        {type === "INFORMATION" && <AlertCircle className="size-6 text-yellow-500" />}
      </div>

      {/* Content */}
      <div className="flex flex-1 flex-col gap-1">
        <div className="flex items-center justify-between gap-2">
          <h3 className="font-semibold text-foreground text-base text-orange-500 font-heading">{heading}</h3>
          {statusCode && (
            <span className={cn(
              "text-[10px] font-mono font-body px-1.5 py-0.5 rounded border whitespace-nowrap shrink-0",
              type === "SUCESS" && "bg-green-500/10 text-green-500 border-green-500/20",
              type === "ERROR" && "bg-red-500/10 text-red-500 border-red-500/20",
              type === "INFORMATION" && "bg-yellow-500/10 text-yellow-500 border-yellow-500/20"
            )}>
              {statusCode}
            </span>
          )}
        </div>
        <div className="text-sm text-muted-foreground break-words">
          <span className="font-medium text-foreground font-body">Message: </span>
          {message}
        </div>
      </div>
    </div>
  ))
}
