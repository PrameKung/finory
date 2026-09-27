"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { LoaderCircleIcon, PencilIcon } from "lucide-react";
import { useState } from "react";
import { useForm } from "react-hook-form";

import { Button } from "@/components/ui/button";
import { InlineMessage } from "@/components/shared/inline-message";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet";
import { WalletFormFields } from "@/features/wallets/components/wallet-form-fields";
import { useUpdateWallet } from "@/features/wallets/hooks/use-update-wallet";
import { updateWalletSchema } from "@/features/wallets/schemas/wallet-schema";
import type {
  UpdateWalletInput,
  Wallet,
} from "@/features/wallets/types/wallet";
import { getApiErrorMessage } from "@/lib/api/client";

function mutationErrorMessage(error: unknown) {
  return getApiErrorMessage(error, {
    defaultMessage: "We could not update this wallet. Please try again.",
    codeMessages: {
      wallet_already_exists: "A wallet with this name already exists.",
      wallet_not_found: "This wallet no longer exists or cannot be edited.",
      invalid_request: "Check the wallet details and try again.",
    },
  });
}

export function EditWalletSheet({ wallet }: { wallet: Wallet }) {
  const [open, setOpen] = useState(false);
  const updateWallet = useUpdateWallet();
  const values = {
    name: wallet.name,
    type: wallet.type,
    balance: wallet.balance,
    currencyCode: wallet.currencyCode,
  };
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<UpdateWalletInput>({
    resolver: zodResolver(updateWalletSchema),
    defaultValues: values,
  });

  function handleOpenChange(nextOpen: boolean) {
    if (!nextOpen && updateWallet.isPending) {
      return;
    }

    setOpen(nextOpen);
    reset(values);
    updateWallet.reset();
  }

  async function onSubmit(input: UpdateWalletInput) {
    try {
      await updateWallet.mutateAsync({ id: wallet.id, input });
      handleOpenChange(false);
    } catch {
      // The mutation error is rendered below the form fields.
    }
  }

  return (
    <Sheet open={open} onOpenChange={handleOpenChange}>
      <SheetTrigger asChild>
        <Button
          type="button"
          variant="ghost"
          size="icon-sm"
          aria-label={`Edit ${wallet.name}`}
        >
          <PencilIcon aria-hidden="true" />
        </Button>
      </SheetTrigger>
      <SheetContent
        className="w-[min(28rem,100vw)] sm:max-w-md"
        closeDisabled={updateWallet.isPending}
      >
        <SheetHeader className="border-b px-5 py-5 pr-14">
          <SheetTitle className="text-lg">Edit wallet</SheetTitle>
          <SheetDescription>
            Update the details and current balance for {wallet.name}.
          </SheetDescription>
        </SheetHeader>

        <form
          className="flex flex-1 flex-col"
          onSubmit={handleSubmit(onSubmit)}
          noValidate
        >
          <div className="flex-1 space-y-6 overflow-y-auto px-5 py-6">
            <WalletFormFields
              idPrefix={`edit-wallet-${wallet.id}`}
              register={register}
              errors={errors}
            />

            {updateWallet.isError ? (
              <InlineMessage>
                {mutationErrorMessage(updateWallet.error)}
              </InlineMessage>
            ) : null}
          </div>

          <SheetFooter className="border-t px-5 py-4 sm:flex-row sm:justify-end">
            <Button
              type="button"
              variant="outline"
              onClick={() => handleOpenChange(false)}
              disabled={updateWallet.isPending}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={updateWallet.isPending}>
              {updateWallet.isPending ? (
                <LoaderCircleIcon
                  className="animate-spin motion-reduce:animate-none"
                  aria-hidden="true"
                />
              ) : null}
              {updateWallet.isPending ? "Saving…" : "Save changes"}
            </Button>
          </SheetFooter>
        </form>
      </SheetContent>
    </Sheet>
  );
}
