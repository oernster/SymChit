// Small pieces of wording used in more than one place.

/** counted writes a count with its noun: "1 event", "3 events". */
export function counted(count: number, noun: string): string {
  return `${count} ${noun}${count === 1 ? '' : 's'}`
}
