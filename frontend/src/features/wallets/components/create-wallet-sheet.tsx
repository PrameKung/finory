"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { LoaderCircleIcon, PlusIcon } from "lucide-react";
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
import { useCreateWallet } from "@/features/wallets/hooks/use-create-wallet";
import { createWalletSchema } from "@/features/wallets/schemas/wallet-schema";
import type { CreateWalletInput } from "@/features/wallets/types/wallet";
import { getApiErrorMessage } from "@/lib/api/client";

function mutationErrorMessage(error: unknown) {
  return getApiErrorMessage(error, {
    defaultMessage: "We could not create this wallet. Please try again.",
    codeMessages: {
      wallet_already_exists: "A wallet with this name already exists.",
      invalid_request: "Check the wallet details and try again.",
    },
  });
}

export function CreateWalletSheet() {
  const [open, setOpen] = useState(false);
  const createWallet = useCreateWallet();
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<CreateWalletInput>({
    resolver: zodResolver(createWalletSchema),
    defaultValues: {
      name: "",
      type: "cash",
      balance: "0",
      currencyCode: "THB",
    },
  });

  function handleOpenChange(nextOpen: boolean) {
    if (!nextOpen && createWallet.isPending) {
      return;
    }

    setOpen(nextOpen);
    if (!nextOpen) {
      reset();
      createWallet.reset();
    }
  }

  async function onSubmit(input: CreateWalletInput) {
    try {
      await createWallet.mutateAsync(input);
      handleOpenChange(false);
    } catch {
      // The mutation error is rendered below the form fields.
    }
  }

  return (
    <Sheet open={open} onOpenChange={handleOpenChange}>
      <SheetTrigger asChild>
        <Button type="button" size="lg">
          <PlusIcon data-icon="inline-start" aria-hidden="true" />
          New wallet
        </Button>
      </SheetTrigger>
      <SheetContent
        className="w-[min(28rem,100vw)] sm:max-w-md"
        closeDisabled={createWallet.isPending}
      >
        <SheetHeader className="border-b px-5 py-5 pr-14">
          <SheetTitle className="text-lg">Create wallet</SheetTitle>
          <SheetDescription>
            Add an account and its current balance.
          </SheetDescription>
        </SheetHeader>

        <form
          className="flex flex-1 flex-col"
          onSubmit={handleSubmit(onSubmit)}
          noValidate
        >
          <div className="flex-1 space-y-6 overflow-y-auto px-5 py-6">
            <WalletFormFields
              idPrefix="create-wallet"
              register={register}
              errors={errors}
            />

            {createWallet.isError ? (
              <InlineMessage>
                {mutationErrorMessage(createWallet.error)}
              </InlineMessage>
            ) : null}
          </div>

          <SheetFooter className="border-t px-5 py-4 sm:flex-row sm:justify-end">
            <Button
              type="button"
              variant="outline"
              onClick={() => handleOpenChange(false)}
              disabled={createWallet.isPending}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={createWallet.isPending}>
              {createWallet.isPending ? (
                <LoaderCircleIcon
                  className="animate-spin motion-reduce:animate-none"
                  aria-hidden="true"
                />
              ) : null}
              {createWallet.isPending ? "Creating…" : "Create wallet"}
            </Button>
          </SheetFooter>
        </form>
      </SheetContent>
    </Sheet>
  );
}
