"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { format } from "date-fns";
import { LoaderCircleIcon, PlusIcon } from "lucide-react";
import Link from "next/link";
import { useEffect, useState } from "react";
import { useForm, useWatch } from "react-hook-form";

import { Button } from "@/components/ui/button";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet";
import { useCategories } from "@/features/categories/hooks/use-categories";
import { TransactionFormFields } from "@/features/transactions/components/transaction-form-fields";
import { useCreateTransaction } from "@/features/transactions/hooks/use-create-transaction";
import { createTransactionSchema } from "@/features/transactions/schemas/transaction-schema";
import type { CreateTransactionInput } from "@/features/transactions/types/transaction";
import { useWallets } from "@/features/wallets/hooks/use-wallets";
import { ApiError } from "@/lib/api/client";

const defaultValues: CreateTransactionInput = {
  categoryId: "",
  walletId: "",
  type: "expense",
  amount: "",
  description: "",
  transactionDate: "",
};

function mutationErrorMessage(error: Error | null) {
  if (
    error instanceof ApiError &&
    error.code === "invalid_transaction_reference"
  ) {
    return "The selected category or wallet is no longer available. Refresh your selections and try again.";
  }

  if (error instanceof ApiError && error.status === 400) {
    return "Check the transaction details and try again.";
  }

  return "We could not add this transaction. Please try again.";
}

export function CreateTransactionSheet() {
  const [open, setOpen] = useState(false);
  const createTransaction = useCreateTransaction();
  const categoriesQuery = useCategories();
  const walletsQuery = useWallets();
  const {
    register,
    handleSubmit,
    reset,
    setValue,
    control,
    formState: { errors },
  } = useForm<CreateTransactionInput>({
    resolver: zodResolver(createTransactionSchema),
    defaultValues,
  });
  const selectedType = useWatch({ control, name: "type" });
  const availableCategories = (categoriesQuery.data ?? []).filter(
    (category) => category.type === selectedType,
  );
  const referencesPending =
    categoriesQuery.isPending || walletsQuery.isPending;
  const referencesError = categoriesQuery.isError || walletsQuery.isError;
  const hasRequiredReferences =
    availableCategories.length > 0 && (walletsQuery.data?.length ?? 0) > 0;

  useEffect(() => {
    setValue("categoryId", "");
  }, [selectedType, setValue]);

  function handleOpenChange(nextOpen: boolean) {
    setOpen(nextOpen);

    if (nextOpen) {
      reset({
        ...defaultValues,
        transactionDate: format(new Date(), "yyyy-MM-dd"),
      });
    } else {
      reset(defaultValues);
      createTransaction.reset();
    }
  }

  async function onSubmit(input: CreateTransactionInput) {
    try {
      await createTransaction.mutateAsync(input);
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
          Add transaction
        </Button>
      </SheetTrigger>
      <SheetContent className="w-[min(32rem,100vw)] sm:max-w-lg">
        <SheetHeader className="border-b px-5 py-5 pr-14">
          <SheetTitle className="text-lg">Add transaction</SheetTitle>
          <SheetDescription>
            Record income or an expense in one of your wallets.
          </SheetDescription>
        </SheetHeader>

        <form
          className="flex flex-1 flex-col overflow-hidden"
          onSubmit={handleSubmit(onSubmit)}
          noValidate
        >
          <div className="flex-1 space-y-6 overflow-y-auto px-5 py-6">
            <TransactionFormFields
              idPrefix="create-transaction"
              register={register}
              errors={errors}
              selectedType={selectedType}
              categories={categoriesQuery.data ?? []}
              wallets={walletsQuery.data ?? []}
              referencesPending={referencesPending}
            />

            {referencesError ? (
              <p
                role="alert"
                className="rounded-lg border border-destructive/20 bg-destructive/5 px-3 py-2.5 text-sm text-destructive"
              >
                Categories or wallets could not be loaded. Close this form and
                try again.
              </p>
            ) : null}

            {!referencesPending && !referencesError && !hasRequiredReferences ? (
              <p className="rounded-lg border bg-muted/40 px-3 py-2.5 text-sm text-muted-foreground">
                You need a {selectedType} category and a wallet before adding
                this transaction. Manage them in{" "}
                <Link href="/categories" className="font-medium text-foreground underline underline-offset-4">
                  Categories
                </Link>{" "}
                and{" "}
                <Link href="/wallets" className="font-medium text-foreground underline underline-offset-4">
                  Wallets
                </Link>
                .
              </p>
            ) : null}

            {createTransaction.isError ? (
              <p
                role="alert"
                className="rounded-lg border border-destructive/20 bg-destructive/5 px-3 py-2.5 text-sm text-destructive"
              >
                {mutationErrorMessage(createTransaction.error)}
              </p>
            ) : null}
          </div>

          <SheetFooter className="border-t px-5 py-4 sm:flex-row sm:justify-end">
            <Button
              type="button"
              variant="outline"
              onClick={() => handleOpenChange(false)}
              disabled={createTransaction.isPending}
            >
              Cancel
            </Button>
            <Button
              type="submit"
              disabled={
                createTransaction.isPending ||
                referencesPending ||
                referencesError ||
                !hasRequiredReferences
              }
            >
              {createTransaction.isPending ? (
                <LoaderCircleIcon
                  className="animate-spin motion-reduce:animate-none"
                  aria-hidden="true"
                />
              ) : null}
              {createTransaction.isPending ? "Adding…" : "Add transaction"}
            </Button>
          </SheetFooter>
        </form>
      </SheetContent>
    </Sheet>
  );
}
