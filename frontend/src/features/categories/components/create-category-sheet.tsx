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
import { CategoryFormFields } from "@/features/categories/components/category-form-fields";
import { createCategorySchema } from "@/features/categories/schemas/category-schema";
import { useCreateCategory } from "@/features/categories/hooks/use-create-category";
import type { CreateCategoryInput } from "@/features/categories/types/category";
import { getApiErrorMessage } from "@/lib/api/client";

function mutationErrorMessage(error: unknown) {
  return getApiErrorMessage(error, {
    defaultMessage: "We could not create this category. Please try again.",
    codeMessages: {
      category_already_exists:
        "A category with this name and type already exists.",
      invalid_request: "Check the category details and try again.",
    },
  });
}

export function CreateCategorySheet() {
  const [open, setOpen] = useState(false);
  const createCategory = useCreateCategory();
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<CreateCategoryInput>({
    resolver: zodResolver(createCategorySchema),
    defaultValues: {
      name: "",
      type: "expense",
    },
  });

  function handleOpenChange(nextOpen: boolean) {
    if (!nextOpen && createCategory.isPending) {
      return;
    }

    setOpen(nextOpen);
    if (!nextOpen) {
      reset();
      createCategory.reset();
    }
  }

  async function onSubmit(input: CreateCategoryInput) {
    try {
      await createCategory.mutateAsync(input);
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
          New category
        </Button>
      </SheetTrigger>
      <SheetContent
        className="w-[min(26rem,100vw)] sm:max-w-md"
        closeDisabled={createCategory.isPending}
      >
        <SheetHeader className="border-b px-5 py-5 pr-14">
          <SheetTitle className="text-lg">Create category</SheetTitle>
          <SheetDescription>
            Add a custom category for organizing your transactions.
          </SheetDescription>
        </SheetHeader>

        <form
          id="create-category-form"
          className="flex flex-1 flex-col"
          onSubmit={handleSubmit(onSubmit)}
          noValidate
        >
          <div className="flex-1 space-y-6 overflow-y-auto px-5 py-6">
            <CategoryFormFields
              nameId="create-category-name"
              register={register}
              errors={errors}
            />

            {createCategory.isError ? (
              <InlineMessage>
                {mutationErrorMessage(createCategory.error)}
              </InlineMessage>
            ) : null}
          </div>

          <SheetFooter className="border-t px-5 py-4 sm:flex-row sm:justify-end">
            <Button
              type="button"
              variant="outline"
              onClick={() => handleOpenChange(false)}
              disabled={createCategory.isPending}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={createCategory.isPending}>
              {createCategory.isPending ? (
                <LoaderCircleIcon
                  className="animate-spin motion-reduce:animate-none"
                  aria-hidden="true"
                />
              ) : null}
              {createCategory.isPending ? "Creating…" : "Create category"}
            </Button>
          </SheetFooter>
        </form>
      </SheetContent>
    </Sheet>
  );
}
