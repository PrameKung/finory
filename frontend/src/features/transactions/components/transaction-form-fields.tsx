import type { FieldErrors, UseFormRegister } from "react-hook-form";

import type { Category } from "@/features/categories/types/category";
import type {
  CreateTransactionInput,
  TransactionType,
} from "@/features/transactions/types/transaction";
import type { Wallet } from "@/features/wallets/types/wallet";
import { cn } from "@/lib/utils";

const transactionTypes = [
  {
    value: "expense",
    label: "Expense",
    description: "Money going out",
  },
  {
    value: "income",
    label: "Income",
    description: "Money coming in",
  },
] as const;

const inputClassName =
  "h-10 w-full rounded-lg border bg-background px-3 text-sm outline-none transition-shadow placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 aria-invalid:border-destructive aria-invalid:ring-3 aria-invalid:ring-destructive/20";

type TransactionFormFieldsProps = {
  idPrefix: string;
  register: UseFormRegister<CreateTransactionInput>;
  errors: FieldErrors<CreateTransactionInput>;
  selectedType: TransactionType;
  categories: Category[];
  wallets: Wallet[];
  referencesPending: boolean;
  onTypeChange?: () => void;
};

function FieldError({ id, message }: { id: string; message?: string }) {
  return message ? (
    <p id={id} className="text-sm text-destructive">
      {message}
    </p>
  ) : null;
}

export function TransactionFormFields({
  idPrefix,
  register,
  errors,
  selectedType,
  categories,
  wallets,
  referencesPending,
  onTypeChange,
}: TransactionFormFieldsProps) {
  const availableCategories = categories.filter(
    (category) => category.type === selectedType,
  );

  return (
    <>
      <fieldset className="space-y-2">
        <legend className="text-sm font-medium">Type</legend>
        <div className="grid grid-cols-2 gap-3">
          {transactionTypes.map((type) => (
            <label
              key={type.value}
              className={cn(
                "has-checked:border-foreground has-checked:bg-muted/60 flex cursor-pointer flex-col rounded-xl border p-3 transition-colors hover:bg-muted/40",
                "has-focus-visible:ring-3 has-focus-visible:ring-ring/50",
              )}
            >
              <input
                type="radio"
                value={type.value}
                className="sr-only"
                {...register("type", { onChange: onTypeChange })}
              />
              <span className="text-sm font-medium">{type.label}</span>
              <span className="text-xs text-muted-foreground">
                {type.description}
              </span>
            </label>
          ))}
        </div>
      </fieldset>

      <div className="space-y-2">
        <label htmlFor={`${idPrefix}-amount`} className="text-sm font-medium">
          Amount
        </label>
        <input
          id={`${idPrefix}-amount`}
          type="text"
          inputMode="decimal"
          autoComplete="off"
          placeholder="0.00"
          aria-invalid={Boolean(errors.amount)}
          aria-describedby={
            errors.amount ? `${idPrefix}-amount-error` : undefined
          }
          className={cn(inputClassName, "tabular-nums")}
          {...register("amount")}
        />
        <FieldError
          id={`${idPrefix}-amount-error`}
          message={errors.amount?.message}
        />
      </div>

      <div className="grid gap-5 sm:grid-cols-2">
        <div className="space-y-2">
          <label
            htmlFor={`${idPrefix}-category`}
            className="text-sm font-medium"
          >
            Category
          </label>
          <select
            id={`${idPrefix}-category`}
            disabled={referencesPending || availableCategories.length === 0}
            aria-invalid={Boolean(errors.categoryId)}
            aria-describedby={
              errors.categoryId ? `${idPrefix}-category-error` : undefined
            }
            className={cn(inputClassName, "disabled:cursor-not-allowed disabled:opacity-50")}
            {...register("categoryId")}
          >
            <option value="">
              {referencesPending
                ? "Loading…"
                : availableCategories.length === 0
                  ? `No ${selectedType} categories`
                  : "Select category"}
            </option>
            {availableCategories.map((category) => (
              <option key={category.id} value={category.id}>
                {category.name}
              </option>
            ))}
          </select>
          <FieldError
            id={`${idPrefix}-category-error`}
            message={errors.categoryId?.message}
          />
        </div>

        <div className="space-y-2">
          <label htmlFor={`${idPrefix}-wallet`} className="text-sm font-medium">
            Wallet
          </label>
          <select
            id={`${idPrefix}-wallet`}
            disabled={referencesPending || wallets.length === 0}
            aria-invalid={Boolean(errors.walletId)}
            aria-describedby={
              errors.walletId ? `${idPrefix}-wallet-error` : undefined
            }
            className={cn(inputClassName, "disabled:cursor-not-allowed disabled:opacity-50")}
            {...register("walletId")}
          >
            <option value="">
              {referencesPending
                ? "Loading…"
                : wallets.length === 0
                  ? "No wallets available"
                  : "Select wallet"}
            </option>
            {wallets.map((wallet) => (
              <option key={wallet.id} value={wallet.id}>
                {wallet.name} ({wallet.currencyCode})
              </option>
            ))}
          </select>
          <FieldError
            id={`${idPrefix}-wallet-error`}
            message={errors.walletId?.message}
          />
        </div>
      </div>

      <div className="space-y-2">
        <label htmlFor={`${idPrefix}-date`} className="text-sm font-medium">
          Date
        </label>
        <input
          id={`${idPrefix}-date`}
          type="date"
          aria-invalid={Boolean(errors.transactionDate)}
          aria-describedby={
            errors.transactionDate ? `${idPrefix}-date-error` : undefined
          }
          className={inputClassName}
          {...register("transactionDate")}
        />
        <FieldError
          id={`${idPrefix}-date-error`}
          message={errors.transactionDate?.message}
        />
      </div>

      <div className="space-y-2">
        <div className="flex items-center justify-between gap-3">
          <label htmlFor={`${idPrefix}-description`} className="text-sm font-medium">
            Note
          </label>
          <span className="text-xs text-muted-foreground">Optional</span>
        </div>
        <textarea
          id={`${idPrefix}-description`}
          rows={3}
          maxLength={500}
          placeholder="What was this transaction for?"
          aria-invalid={Boolean(errors.description)}
          aria-describedby={
            errors.description ? `${idPrefix}-description-error` : undefined
          }
          className="min-h-24 w-full resize-y rounded-lg border bg-background px-3 py-2.5 text-sm outline-none transition-shadow placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 aria-invalid:border-destructive aria-invalid:ring-3 aria-invalid:ring-destructive/20"
          {...register("description")}
        />
        <FieldError
          id={`${idPrefix}-description-error`}
          message={errors.description?.message}
        />
      </div>
    </>
  );
}
