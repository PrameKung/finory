import { LoadingState } from "@/components/shared/loading-state";

export default function Loading() {
  return (
    <LoadingState
      title="Loading page"
      description="Your latest financial information is on the way."
    />
  );
}
