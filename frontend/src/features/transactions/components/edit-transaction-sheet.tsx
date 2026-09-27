"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { LoaderCircleIcon, PencilIcon } from "lucide-react";
import Link from "next/link";
import { useState } from "react";
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
import { useUpdateTransaction } from "@/features/transactions/hooks/use-update-transaction";
import { updateTransactionSchema } from "@/features/transactions/schemas/transaction-schema";
import type {
  Transaction,
  UpdateTransactionInput,
} from "@/features/transactions/types/transaction";
import { useWallets } from "@/features/wallets/hooks/use-wallets";
import { ApiError } from "@/lib/api/client";

function transactionValues(transaction: Transaction): UpdateTransactionInput {
  return {
    categoryId: transaction.categoryId,
    walletId: transaction.walletId,
    type: transaction.type,
    amount: transaction.amount,
    description: transaction.description ?? "",
    transactionDate: transaction.transactionDate,
  };
}

function mutationErrorMessage(error: Error | null) {
  if (error instanceof ApiError && error.status === 404) {
    return "This transaction no longer exists or cannot be edited.";
  }

  if (
    error instanceof ApiError &&
    error.code === "invalid_transaction_reference"
  ) {
    return "The selected category or wallet is no longer available. Refresh your selections and try again.";
  }

  if (error instanceof ApiError && error.status === 400) {
    return "Check the transaction details and try again.";
  }

  return "We could not update this transaction. Please try again.";
}

export function EditTransactionSheet({
  transaction,
}: {
  transaction: Transaction;
}) {
  const [open, setOpen] = useState(false);
  const updateTransaction = useUpdateTransaction();
  const categoriesQuery = useCategories();
  const walletsQuery = useWallets();
  const initialValues = transactionValues(transaction);
  const {
    register,
    handleSubmit,
    reset,
    setValue,
    control,
    formState: { errors },
  } = useForm<UpdateTransactionInput>({
    resolver: zodResolver(updateTransactionSchema),
    defaultValues: initialValues,
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

  function handleOpenChange(nextOpen: boolean) {
    setOpen(nextOpen);
    reset(transactionValues(transaction));
    updateTransaction.reset();
  }

  async function onSubmit(input: UpdateTransactionInput) {
    try {
      await updateTransaction.mutateAsync({ id: transaction.id, input });
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
          aria-label={`Edit ${transaction.description || "transaction"}`}
        >
          <PencilIcon aria-hidden="true" />
        </Button>
      </SheetTrigger>
      <SheetContent className="w-[min(32rem,100vw)] sm:max-w-lg">
        <SheetHeader className="border-b px-5 py-5 pr-14">
          <SheetTitle className="text-lg">Edit transaction</SheetTitle>
          <SheetDescription>
            Update this transaction&apos;s amount, category, wallet, or date.
          </SheetDescription>
        </SheetHeader>

        <form
          className="flex flex-1 flex-col overflow-hidden"
          onSubmit={handleSubmit(onSubmit)}
          noValidate
        >
          <div className="flex-1 space-y-6 overflow-y-auto px-5 py-6">
            <TransactionFormFields
              idPrefix={`edit-transaction-${transaction.id}`}
              register={register}
              errors={errors}
              selectedType={selectedType}
              categories={categoriesQuery.data ?? []}
              wallets={walletsQuery.data ?? []}
              referencesPending={referencesPending}
              onTypeChange={() => setValue("categoryId", "")}
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
                You need a {selectedType} category and a wallet to save this
                transaction. Manage them in{" "}
                <Link
                  href="/categories"
                  className="font-medium text-foreground underline underline-offset-4"
                >
                  Categories
                </Link>{" "}
                and{" "}
                <Link
                  href="/wallets"
                  className="font-medium text-foreground underline underline-offset-4"
                >
                  Wallets
                </Link>
                .
              </p>
            ) : null}

            {updateTransaction.isError ? (
              <p
                role="alert"
                className="rounded-lg border border-destructive/20 bg-destructive/5 px-3 py-2.5 text-sm text-destructive"
              >
                {mutationErrorMessage(updateTransaction.error)}
              </p>
            ) : null}
          </div>

          <SheetFooter className="border-t px-5 py-4 sm:flex-row sm:justify-end">
            <Button
              type="button"
              variant="outline"
              onClick={() => handleOpenChange(false)}
              disabled={updateTransaction.isPending}
            >
              Cancel
            </Button>
            <Button
              type="submit"
              disabled={
                updateTransaction.isPending ||
                referencesPending ||
                referencesError ||
                !hasRequiredReferences
              }
            >
              {updateTransaction.isPending ? (
                <LoaderCircleIcon
                  className="animate-spin motion-reduce:animate-none"
                  aria-hidden="true"
                />
              ) : null}
              {updateTransaction.isPending ? "Saving…" : "Save changes"}
            </Button>
          </SheetFooter>
        </form>
      </SheetContent>
    </Sheet>
  );
}
