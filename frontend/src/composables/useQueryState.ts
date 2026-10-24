import { ref, watch, type Ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

/**
 * 列表页状态 URL 化（设计审计 D19 / P1-02）。
 *
 * 把组件内存态（筛选 / 分页 / Tab / 排序）同步到 route.query：
 * - 读：初始化时从 query 解析，值非法/缺省回落 default
 * - 写：router.replace（不堆积浏览器历史记录）
 * - 前进/后退：route.query 变化时回写 state
 *
 * 用法：
 *   const page = useQueryState('page', 1)      // Ref<number>
 *   const keyword = useQueryState('kw', '')    // Ref<string>
 *   const tab = useQueryState('tab', 'running') // Ref<string>
 *
 * 值等于 default 时从 query 中省略，保持 URL 短且默认视图链接干净。
 *
 * 类型说明：字面量默认值（1 / '' / '/'）会被 TS 推断为字面量类型（1 / "" / "/"），
 * 导致后续赋值报错；此处统一放宽为 string / number。
 */
type WidenLiteral<T> = T extends string ? string : T extends number ? number : T

export function useQueryState<T extends string | number>(key: string, defaultValue: T): Ref<WidenLiteral<T>> {
  const route = useRoute()
  const router = useRouter()

  function parse(raw: unknown): WidenLiteral<T> {
    if (raw === undefined || raw === null || raw === '') return defaultValue as WidenLiteral<T>
    if (typeof defaultValue === 'number') {
      const n = Number(raw)
      return (Number.isFinite(n) && n > 0 ? n : defaultValue) as WidenLiteral<T>
    }
    return raw as WidenLiteral<T>
  }

  const state = ref<WidenLiteral<T>>(parse(route.query[key])) as Ref<WidenLiteral<T>>

  // state → URL（replace，防历史堆积；等值时不触发导航）
  // 运行时防御：NSelect clearable 等组件可能把 null 写进 state，视同默认值从 URL 省略
  watch(state, (val) => {
    const raw = val as unknown
    const next = raw === defaultValue || raw === '' || raw === null || raw === undefined ? undefined : String(raw)
    const current = route.query[key]
    if ((current ?? undefined) === next) return
    router.replace({ query: { ...route.query, [key]: next } })
  })

  // URL → state（浏览器前进/后退、外部改链接）
  watch(
    () => route.query[key],
    (raw) => {
      const parsed = parse(raw)
      if (parsed !== state.value) state.value = parsed
    }
  )

  return state
}
