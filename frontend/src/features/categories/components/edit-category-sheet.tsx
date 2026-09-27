"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { LoaderCircleIcon, PencilIcon } from "lucide-react";
import { useState } from "react";
import { useForm } from "react-hook-form";

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
import { CategoryFormFields } from "@/features/categories/components/category-form-fields";
import { useUpdateCategory } from "@/features/categories/hooks/use-update-category";
import { updateCategorySchema } from "@/features/categories/schemas/category-schema";
import type {
  Category,
  UpdateCategoryInput,
} from "@/features/categories/types/category";
import { getApiErrorMessage } from "@/lib/api/client";

function mutationErrorMessage(error: unknown) {
  return getApiErrorMessage(error, {
    defaultMessage: "We could not update this category. Please try again.",
    codeMessages: {
      category_already_exists:
        "A category with this name and type already exists.",
      category_not_found: "This category no longer exists or cannot be edited.",
      invalid_request: "Check the category details and try again.",
    },
  });
}

export function EditCategorySheet({ category }: { category: Category }) {
  const [open, setOpen] = useState(false);
  const updateCategory = useUpdateCategory();
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<UpdateCategoryInput>({
    resolver: zodResolver(updateCategorySchema),
    defaultValues: {
      name: category.name,
      type: category.type,
    },
  });

  function handleOpenChange(nextOpen: boolean) {
    setOpen(nextOpen);
    reset({ name: category.name, type: category.type });
    updateCategory.reset();
  }

  async function onSubmit(input: UpdateCategoryInput) {
    try {
      await updateCategory.mutateAsync({ id: category.id, input });
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
          aria-label={`Edit ${category.name}`}
        >
          <PencilIcon aria-hidden="true" />
        </Button>
      </SheetTrigger>
      <SheetContent className="w-[min(26rem,100vw)] sm:max-w-md">
        <SheetHeader className="border-b px-5 py-5 pr-14">
          <SheetTitle className="text-lg">Edit category</SheetTitle>
          <SheetDescription>
            Update the name or type for {category.name}.
          </SheetDescription>
        </SheetHeader>

        <form
          className="flex flex-1 flex-col"
          onSubmit={handleSubmit(onSubmit)}
          noValidate
        >
          <div className="flex-1 space-y-6 overflow-y-auto px-5 py-6">
            <CategoryFormFields
              nameId={`edit-category-name-${category.id}`}
              register={register}
              errors={errors}
            />

            {updateCategory.isError ? (
              <p
                role="alert"
                className="rounded-lg border border-destructive/20 bg-destructive/5 px-3 py-2.5 text-sm text-destructive"
              >
                {mutationErrorMessage(updateCategory.error)}
              </p>
            ) : null}
          </div>

          <SheetFooter className="border-t px-5 py-4 sm:flex-row sm:justify-end">
            <Button
              type="button"
              variant="outline"
              onClick={() => handleOpenChange(false)}
              disabled={updateCategory.isPending}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={updateCategory.isPending}>
              {updateCategory.isPending ? (
                <LoaderCircleIcon
                  className="animate-spin motion-reduce:animate-none"
                  aria-hidden="true"
                />
              ) : null}
              {updateCategory.isPending ? "Saving…" : "Save changes"}
            </Button>
          </SheetFooter>
        </form>
      </SheetContent>
    </Sheet>
  );
}
