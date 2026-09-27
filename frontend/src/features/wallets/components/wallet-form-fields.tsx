import type { FieldErrors, UseFormRegister } from "react-hook-form";

import type { CreateWalletInput } from "@/features/wallets/types/wallet";
import { cn } from "@/lib/utils";

const walletTypes = [
  { value: "cash", label: "Cash" },
  { value: "bank", label: "Bank account" },
  { value: "e_wallet", label: "E-wallet" },
  { value: "other", label: "Other" },
] as const;

type WalletFormFieldsProps = {
  idPrefix: string;
  register: UseFormRegister<CreateWalletInput>;
  errors: FieldErrors<CreateWalletInput>;
};

export function WalletFormFields({
  idPrefix,
  register,
  errors,
}: WalletFormFieldsProps) {
  return (
    <>
      <div className="space-y-2">
        <label htmlFor={`${idPrefix}-name`} className="text-sm font-medium">
          Name
        </label>
        <input
          id={`${idPrefix}-name`}
          type="text"
          autoComplete="off"
          maxLength={100}
          placeholder="e.g. Everyday account"
          aria-invalid={Boolean(errors.name)}
          aria-describedby={errors.name ? `${idPrefix}-name-error` : undefined}
          className="h-10 w-full rounded-lg border bg-background px-3 text-sm outline-none transition-shadow placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 aria-invalid:border-destructive aria-invalid:ring-3 aria-invalid:ring-destructive/20"
          {...register("name")}
        />
        {errors.name ? (
          <p id={`${idPrefix}-name-error`} className="text-sm text-destructive">
            {errors.name.message}
          </p>
        ) : null}
      </div>

      <fieldset className="space-y-2">
        <legend className="text-sm font-medium">Wallet type</legend>
        <div className="grid grid-cols-2 gap-3">
          {walletTypes.map((type) => (
            <label
              key={type.value}
              className={cn(
                "has-checked:border-foreground has-checked:bg-muted/60 flex cursor-pointer rounded-xl border p-3 text-sm font-medium transition-colors hover:bg-muted/40",
                "has-focus-visible:ring-3 has-focus-visible:ring-ring/50",
              )}
            >
              <input
                type="radio"
                value={type.value}
                className="sr-only"
                {...register("type")}
              />
              {type.label}
            </label>
          ))}
        </div>
      </fieldset>

      <div className="grid grid-cols-[minmax(0,1fr)_7rem] gap-3">
        <div className="space-y-2">
          <label htmlFor={`${idPrefix}-balance`} className="text-sm font-medium">
            Balance
          </label>
          <input
            id={`${idPrefix}-balance`}
            type="text"
            inputMode="decimal"
            autoComplete="off"
            placeholder="0.00"
            aria-invalid={Boolean(errors.balance)}
            aria-describedby={
              errors.balance ? `${idPrefix}-balance-error` : undefined
            }
            className="h-10 w-full rounded-lg border bg-background px-3 text-sm tabular-nums outline-none transition-shadow placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 aria-invalid:border-destructive aria-invalid:ring-3 aria-invalid:ring-destructive/20"
            {...register("balance")}
          />
        </div>
        <div className="space-y-2">
          <label htmlFor={`${idPrefix}-currency`} className="text-sm font-medium">
            Currency
          </label>
          <input
            id={`${idPrefix}-currency`}
            type="text"
            autoComplete="off"
            maxLength={3}
            placeholder="THB"
            aria-invalid={Boolean(errors.currencyCode)}
            aria-describedby={
              errors.currencyCode ? `${idPrefix}-currency-error` : undefined
            }
            className="h-10 w-full rounded-lg border bg-background px-3 text-sm uppercase outline-none transition-shadow placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 aria-invalid:border-destructive aria-invalid:ring-3 aria-invalid:ring-destructive/20"
            {...register("currencyCode", {
              setValueAs: (value: unknown) =>
                typeof value === "string"
                  ? value.trim().toUpperCase()
                  : value,
            })}
          />
        </div>
      </div>
      {errors.balance || errors.currencyCode ? (
        <div className="space-y-1">
          {errors.balance ? (
            <p
              id={`${idPrefix}-balance-error`}
              className="text-sm text-destructive"
            >
              {errors.balance.message}
            </p>
          ) : null}
          {errors.currencyCode ? (
            <p
              id={`${idPrefix}-currency-error`}
              className="text-sm text-destructive"
            >
              {errors.currencyCode.message}
            </p>
          ) : null}
        </div>
      ) : null}
    </>
  );
}
