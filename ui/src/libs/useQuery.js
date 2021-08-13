export default function useQuery() {
  return new URLSearchParams(useLocation().search);
}
