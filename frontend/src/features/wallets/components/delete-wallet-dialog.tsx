"use client";

import { LoaderCircleIcon, Trash2Icon } from "lucide-react";
import { useState } from "react";

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog";
import { Button } from "@/components/ui/button";
import { useDeleteWallet } from "@/features/wallets/hooks/use-delete-wallet";
import type { Wallet } from "@/features/wallets/types/wallet";

export function DeleteWalletDialog({ wallet }: { wallet: Wallet }) {
  const [open, setOpen] = useState(false);
  const deleteWallet = useDeleteWallet();

  function handleOpenChange(nextOpen: boolean) {
    setOpen(nextOpen);
    if (!nextOpen) {
      deleteWallet.reset();
    }
  }

  async function handleDelete() {
    try {
      await deleteWallet.mutateAsync(wallet.id);
      handleOpenChange(false);
    } catch {
      // The mutation error is rendered in the confirmation dialog.
    }
  }

  return (
    <AlertDialog open={open} onOpenChange={handleOpenChange}>
      <AlertDialogTrigger asChild>
        <Button
          type="button"
          variant="ghost"
          size="icon-sm"
          aria-label={`Delete ${wallet.name}`}
          className="text-muted-foreground hover:text-destructive"
        >
          <Trash2Icon aria-hidden="true" />
        </Button>
      </AlertDialogTrigger>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Delete {wallet.name}?</AlertDialogTitle>
          <AlertDialogDescription>
            This permanently removes the wallet. This action cannot be undone.
          </AlertDialogDescription>
        </AlertDialogHeader>

        {deleteWallet.isError ? (
          <p
            role="alert"
            className="mt-4 rounded-lg border border-destructive/20 bg-destructive/5 px-3 py-2.5 text-sm text-destructive"
          >
            We could not delete this wallet. It may be in use by a transaction.
          </p>
        ) : null}

        <AlertDialogFooter>
          <AlertDialogCancel asChild>
            <Button variant="outline" disabled={deleteWallet.isPending}>
              Cancel
            </Button>
          </AlertDialogCancel>
          <AlertDialogAction asChild>
            <Button
              variant="destructive"
              disabled={deleteWallet.isPending}
              onClick={(event) => {
                event.preventDefault();
                void handleDelete();
              }}
            >
              {deleteWallet.isPending ? (
                <LoaderCircleIcon
                  className="animate-spin motion-reduce:animate-none"
                  aria-hidden="true"
                />
              ) : null}
              {deleteWallet.isPending ? "Deleting…" : "Delete wallet"}
            </Button>
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
