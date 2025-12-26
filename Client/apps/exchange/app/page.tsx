import { Button } from "@workspace/ui/components/button"

export default function Page() {
  return (
    <div className="flex flex-col items-center justify-center min-h-svh py-10">
      <div className="flex flex-col items-center justify-center gap-4 mb-10">
        <h1 className="text-2xl font-bold">Hello World</h1>
        <Button size="sm">Button</Button>
      </div>

      {/* Dummy content to enable scrolling */}
      <div className="space-y-4 w-full max-w-md px-4">
        {Array.from({ length: 20 }).map((_, i) => (
          <div key={i} className="p-6 border rounded-lg shadow-sm bg-card text-card-foreground">
            <h3 className="font-semibold mb-2">Section {i + 1}</h3>
            <p className="text-muted-foreground">
              This is some dummy content to make the page scrollable.
              Scroll down to hide the header, and scroll up to show it again.
            </p>
          </div>
        ))}
      </div>
    </div>
  )
}
